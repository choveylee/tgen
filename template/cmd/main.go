// Command {{app_name}} starts the generated HTTP service.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"
	"github.com/choveylee/tserver"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "{{domain}}/{{app_name}}/cmd/init"

	"{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/cron"
	"{{domain}}/{{app_name}}/internal/lib"
	"{{domain}}/{{app_name}}/internal/model"
	"{{domain}}/{{app_name}}/internal/router"
	"{{domain}}/{{app_name}}/internal/service"
)

// main validates startup settings, initializes dependencies and serves HTTP.
func main() {
	ctx := context.Background()

	httpPort := tcfg.DefaultInt(tcfg.LocalKey("HTTP_PORT"), 8080)
	if httpPort < 1 || httpPort > 65535 {
		tlog.E(ctx).Msgf("Start HTTP server err (HTTP port %d must be between 1 and 65535)", httpPort)

		os.Exit(1)
	}

	// Validate library configuration before applying migrations.
	errx := lib.InitLib(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("Main err (init lib)")

		os.Exit(1)
	}

	errx = runMigrate(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("Main err (run migrate)")

		os.Exit(1)
	}

	// init model
	errx = model.InitModel(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("Main err (init model)")

		os.Exit(1)
	}

	// init cron
	errx = crontab.InitCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("Main err (init cron)")

		os.Exit(1)
	}

	errx = crontab.StartCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("Main err (start cron)")

		os.Exit(1)
	}

	// init service
	errx = service.InitService(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("Main err (init service)")

		os.Exit(1)
	}

	probeCtx, cancelProbe := context.WithTimeout(ctx, 30*time.Second)
	defer cancelProbe()
	go func() {
		if err := waitForTCPDial(probeCtx, httpPort); err != nil {
			tlog.W(ctx).Err(err).Msg("Main err (wait for http listener)")

			return
		}

		errx := lib.PingServer(probeCtx)
		if errx != nil {
			tlog.W(ctx).Err(errx).Msg("Main err (ping server)")
		} else {
			tlog.I(ctx).Msg("Start http server info (ready)")
		}
	}()

	if err := tserver.StartHttpServer(ctx, router.NewRouter(ctx), httpPort); err != nil {
		tlog.F(ctx).Err(err).Msg("Main err (start http server)")
	}
}

// runMigrate reads the service DSN and applies the bundled migration files.
func runMigrate(ctx context.Context) *terror.Terror {
	runMode := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug))
	serverDSN := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey(runMode+"::SERVER_MYSQL_DSN"), ""))
	config, err := mysql.ParseDSN(serverDSN)
	if err != nil || config.DBName == "" {
		errMsg := tlog.E(ctx).Msgf("Apply migrations (run mode: %s) err (server mysql dsn must specify a database)",
			runMode)

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server mysql dsn"), errMsg)

		return errx
	}

	if err := applyMigrations(ctx, config); err != nil {
		errMsg := tlog.E(ctx).Detailf("error_type:%T", err).Msgf("Apply migrations (run mode: %s) err (apply mysql migrations)",
			runMode)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	return nil
}

// applyMigrations opens the configured database and creates it only for the
// MySQL unknown-database error. Driver and pool ownership are kept explicit.
func applyMigrations(ctx context.Context, config *mysql.Config) (err error) {
	timeout := tcfg.DefaultDuration(tcfg.LocalKey("MYSQL_MIGRATION_TIMEOUT"), 5*time.Minute)
	if timeout <= 0 {
		return errors.New("mysql migration timeout must be positive")
	}

	ctx, cancelMigration := context.WithTimeout(ctx, timeout)
	defer cancelMigration()
	defer func() {
		if ctx.Err() != nil {
			err = fmt.Errorf("run database migrations: %w", ctx.Err())
		}
	}()

	migrationConfig := config.Clone()
	migrationConfig.MultiStatements = true
	originalDial := migrationConfig.DialFunc
	migrationConfig.DialFunc = func(dialCtx context.Context, network, address string) (net.Conn, error) {
		dial := originalDial
		if dial == nil {
			dialer := &net.Dialer{}
			dial = dialer.DialContext
		}

		connection, err := dial(dialCtx, network, address)
		if err != nil {
			return nil, err
		}

		// The migration driver uses Background for SQL. Closing its socket interrupts
		// all statements and lock waits when the caller cancels or the budget expires.
		stopCancel := context.AfterFunc(ctx, func() { _ = connection.Close() })
		tracked := &migrationConnection{Conn: connection, stopCancel: stopCancel}

		return tracked, nil
	}

	connector, err := mysql.NewConnector(migrationConfig)
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}

	database := sql.OpenDB(connector)

	defer database.Close()

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	connection, err := database.Conn(connectCtx)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1049 {
		if err := createMigrationDatabase(connectCtx, config); err != nil {
			return fmt.Errorf("create migration database: %w", err)
		}

		connection, err = database.Conn(connectCtx)
	}

	if err != nil {
		return fmt.Errorf("connect to migration database: %w", err)
	}

	driver, err := migratemysql.WithConnection(connectCtx, connection, &migratemysql.Config{DatabaseName: config.DBName})
	if err != nil {
		_ = connection.Close()

		return fmt.Errorf("initialize migration driver: %w", err)
	}

	client, err := migrate.NewWithDatabaseInstance("file://migration", "mysql", driver)
	if err != nil {
		_ = driver.Close()

		return fmt.Errorf("open migration files: %w", err)
	}

	defer func() {
		if err := errors.Join(client.Close()); err != nil {
			tlog.W(ctx).Detailf("error_type:%T", err).Msg("Close migration client err")
		}
	}()

	err = client.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run database migrations: %w", err)
	}

	return nil
}

// migrationConnection releases its cancellation callback when the driver closes the socket.
type migrationConnection struct {
	net.Conn
	stopCancel func() bool
}

func (connection *migrationConnection) Close() error {
	connection.stopCancel()

	return connection.Conn.Close()
}

// createMigrationDatabase quotes the configured database name as an identifier.
func createMigrationDatabase(ctx context.Context, config *mysql.Config) error {
	adminConfig := config.Clone()
	adminConfig.DBName = ""
	database, err := sql.Open("mysql", adminConfig.FormatDSN())
	if err != nil {
		return err
	}

	defer database.Close()

	databaseName := "`" + strings.ReplaceAll(config.DBName, "`", "``") + "`"
	_, err = database.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+databaseName+" DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")

	return err
}

// waitForTCPDial waits for the local listener until the health probe expires.
func waitForTCPDial(ctx context.Context, port int) error {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	dialer := net.Dialer{Timeout: 150 * time.Millisecond}
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		connection, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			_ = connection.Close()

			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
