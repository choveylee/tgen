// Command {{app_name}} starts the generated HTTP service.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/terror"
	"github.com/choveylee/thttp"
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

func main() {
	ctx := context.Background()

	httpPort := tcfg.DefaultInt(tcfg.LocalKey("HTTP_PORT"), 8080)
	if httpPort < 1 || httpPort > 65535 {
		tlog.E(ctx).Msgf("Start HTTP server err (HTTP port %d must be between 1 and 65535)", httpPort)

		os.Exit(1)
	}

	// init migrations
	errx := runMigrate(ctx)
	if errx != nil {
		tlog.E(ctx).Msg("Service startup stopped during database migration")

		os.Exit(1)
	}

	// init lib
	errx = lib.InitLib(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during shared library initialization")

		os.Exit(1)
	}

	// init model
	errx = model.InitModel(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during model initialization")

		os.Exit(1)
	}

	// init cron
	errx = crontab.InitCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during cron configuration initialization")

		os.Exit(1)
	}

	errx = crontab.StartCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during cron job registration")

		os.Exit(1)
	}

	// init service
	errx = service.InitService(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during service initialization")

		os.Exit(1)
	}

	probeCtx, cancelProbe := context.WithTimeout(ctx, 30*time.Second)
	defer cancelProbe()
	go func() {
		if err := waitForTCPDial(probeCtx, httpPort); err != nil {
			tlog.W(ctx).Msg("startup health probe was skipped because the HTTP listener did not become ready before the timeout elapsed")

			return
		}

		errx := pingServer(probeCtx, httpPort)
		if errx != nil {
			tlog.W(ctx).Msg("startup health probe failed after the HTTP listener became ready")
		} else {
			tlog.I(ctx).Msg("HTTP server started successfully")
		}
	}()

	if err := tserver.StartHttpServer(ctx, router.NewRouter(ctx), httpPort); err != nil {
		tlog.F(ctx).Err(err).Msg("HTTP server exited with an error")
	}
}

// runMigrate reads the service DSN and applies the bundled migration files.
func runMigrate(ctx context.Context) *terror.Terror {
	runMode := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug))
	serverDSN := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey(runMode+"::SERVER_MYSQL_DSN"), ""))
	config, err := mysql.ParseDSN(serverDSN)
	if err != nil || config.DBName == "" {
		errMsg := tlog.E(ctx).Msgf("Apply migrations err (server mysql dsn must specify a database, run_mode=%s)", runMode)

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server mysql dsn"), errMsg)

		return errx
	}

	if err := applyMigrations(ctx, config); err != nil {
		errMsg := tlog.E(ctx).Detailf("error_type:%T", err).Msgf("Apply MySQL migrations err (run_mode=%s)", runMode)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	return nil
}

// applyMigrations opens the configured database and creates it only for the
// MySQL unknown-database error. Driver and pool ownership are kept explicit.
func applyMigrations(ctx context.Context, config *mysql.Config) error {
	migrationConfig := config.Clone()
	migrationConfig.MultiStatements = true
	database, err := sql.Open("mysql", migrationConfig.FormatDSN())
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}

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
		sourceErr, databaseErr := client.Close()
		if sourceErr != nil || databaseErr != nil {
			tlog.W(ctx).Detailf("source_error_type:%T", sourceErr).Detailf("database_error_type:%T", databaseErr).Msg("Close migration client err")
		}
	}()

	if err := client.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run database migrations: %w", err)
	}

	return nil
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

// resolvePingURL attaches the health path and uses HTTP_PORT for loopback
// targets. A host without a scheme uses HTTP.
func resolvePingURL(httpPort int) (*url.URL, error) {
	serverPingHost := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("SERVER_PING_HOST"), ""))
	if serverPingHost == "" {
		serverPingHost = "http://127.0.0.1"
	} else if !strings.Contains(serverPingHost, "://") {
		serverPingHost = "http://" + serverPingHost
	}

	parsedURL, err := url.Parse(serverPingHost)
	if err != nil || parsedURL.Hostname() == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, errors.New("server ping host must be a valid HTTP or HTTPS URL")
	}

	switch parsedURL.Hostname() {
	case "127.0.0.1", "localhost", "::1":
		parsedURL.Host = net.JoinHostPort(parsedURL.Hostname(), fmt.Sprint(httpPort))
	}

	return parsedURL.JoinPath("healthz"), nil
}

// pingServer checks the health endpoint within the startup probe deadline.
func pingServer(ctx context.Context, httpPort int) *terror.Terror {
	pingCount := tcfg.DefaultInt(tcfg.LocalKey("SERVER_PING_COUNT"), 3)
	pingURL, err := resolvePingURL(httpPort)
	if err != nil {
		errMsg := tlog.E(ctx).Msg("Check startup health err (server ping host invalid)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server ping host"), errMsg)

		return errx
	}

	for attempt := range pingCount {
		if attempt > 0 {
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
			case <-timer.C:
			}
		}

		if ctx.Err() != nil {
			break
		}

		response, err := thttp.Get(ctx, pingURL.String(), nil, nil)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
	}

	errMsg := tlog.E(ctx).Msg("Startup health probe failed after all retry attempts")

	errx := terror.NewRawTerror(ctx, terror.ErrSvcExecute("http server"), errMsg)

	return errx
}
