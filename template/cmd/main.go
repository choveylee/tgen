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
	"strconv"
	"strings"
	"time"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/terror"
	"github.com/choveylee/thttp"
	"github.com/choveylee/tlog"
	"github.com/choveylee/tserver"
	"github.com/choveylee/tutil"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
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

	// init migrations
	errx := runMigrate(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during database migration")

		return
	}

	// init lib
	errx = lib.InitLib(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during shared library initialization")

		return
	}

	// init model
	errx = model.InitModel(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during model initialization")

		return
	}

	// init cron
	errx = crontab.InitCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during cron configuration initialization")

		return
	}

	errx = crontab.StartCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during cron job registration")

		return
	}

	// init service
	errx = service.InitService(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("service startup failed during service initialization")

		return
	}

	httpPort := tcfg.DefaultInt(tcfg.LocalKey("HTTP_PORT"), 8080)

	go func() {
		if err := waitForTcpDial(ctx, httpPort, 30*time.Second); err != nil {
			tlog.W(ctx).Msg("startup health probe was skipped because the HTTP listener did not become ready before the timeout elapsed")

			return
		}

		errx := pingServer(ctx, httpPort)
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

func runMigrate(ctx context.Context) *terror.Terror {
	runMode := tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug)

	serverDsn, err := tcfg.String(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_MYSQL_DSN")))
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("database migration setup failed while reading configuration key %q (run_mode=%s)",
			fmt.Sprintf("%s::%s", runMode, "SERVER_MYSQL_DSN"), runMode,
		)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	serverDBName := ""
	if parsedCfg, parseErr := mysql.ParseDSN(serverDsn); parseErr == nil {
		serverDBName = parsedCfg.DBName
	}

	client, err := migrate.New("file://migration", "mysql://"+tutil.MysqlDsnEncode(serverDsn))
	if err != nil {
		initialClientErr := err

		serverCfg, err := mysql.ParseDSN(serverDsn)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("database migration setup failed while parsing the MySQL DSN (run_mode=%s, migration_source=file://migration, initial_migration_client_error=%v)",
				runMode, initialClientErr,
			)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		dbName := serverCfg.DBName

		tlog.W(ctx).Err(initialClientErr).Msgf("database migration client creation failed; attempting database creation fallback (run_mode=%s, database=%q, migration_source=file://migration)",
			runMode, dbName,
		)

		serverCfg.DBName = ""
		tmpDsn := serverCfg.FormatDSN()

		db, err := sql.Open("mysql", tmpDsn)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("database migration setup failed while opening the MySQL connection (run_mode=%s, database=%q, migration_source=file://migration, initial_migration_client_error=%v)",
				runMode, dbName, initialClientErr,
			)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		defer db.Close()

		_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + "`" + dbName + "` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("database migration setup failed while creating database %q (run_mode=%s, migration_source=file://migration, initial_migration_client_error=%v)",
				dbName, runMode, initialClientErr,
			)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		client, err = migrate.New("file://migration", "mysql://"+tutil.MysqlDsnEncode(serverDsn))
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("database migration setup failed while creating the migration client (run_mode=%s, database=%q, migration_source=file://migration, initial_migration_client_error=%v)",
				runMode, dbName, initialClientErr,
			)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}
	}

	defer func() {
		srcErr, dbErr := client.Close()
		if srcErr != nil || dbErr != nil {
			tlog.W(ctx).Msgf("database migration cleanup reported errors (source: %v, database: %v)", srcErr, dbErr)
		}
	}()

	err = client.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		if serverDBName != "" {
			errMsg := tlog.E(ctx).Err(err).Msgf("database migration failed while applying migration files (run_mode=%s, database=%q, migration_source=file://migration)",
				runMode, serverDBName,
			)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		errMsg := tlog.E(ctx).Err(err).Msgf("database migration failed while applying migration files (run_mode=%s, migration_source=file://migration)",
			runMode,
		)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	return nil
}

func waitForTcpDial(ctx context.Context, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	address := fmt.Sprintf("127.0.0.1:%d", port)

	dialer := net.Dialer{Timeout: 150 * time.Millisecond}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			_ = conn.Close()

			return nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("timed out waiting for the TCP listener at %s", address)
}

// resolvePingBaseUrl aligns the health probe base URL with the configured HTTP port.
// When the configured host is empty or points to a loopback address, the function uses
// httpPort so that changing HTTP_PORT does not require a separate ping-host update.
func resolvePingBaseUrl(httpPort int) string {
	serverPingHost := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("SERVER_PING_HOST"), ""))
	if serverPingHost == "" {
		return fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	}

	parsedUrl, err := url.Parse(serverPingHost)
	if err != nil {
		return fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	}

	if parsedUrl.Scheme == "" {
		parsedUrl.Scheme = "http"
	}

	switch parsedUrl.Hostname() {
	case "127.0.0.1", "localhost", "::1":
		parsedUrl.Host = net.JoinHostPort(parsedUrl.Hostname(), strconv.Itoa(httpPort))
	}

	return strings.TrimSuffix(parsedUrl.String(), "/")
}

func pingServer(ctx context.Context, httpPort int) *terror.Terror {
	pingCount := tcfg.DefaultInt(tcfg.LocalKey("SERVER_PING_COUNT"), 3)

	baseUrl := resolvePingBaseUrl(httpPort)
	pingUrl := baseUrl + "/healthz"

	for i := 0; i < pingCount; i++ {
		if i > 0 {
			time.Sleep(time.Second)
		}

		response, err := thttp.Get(ctx, pingUrl, nil, nil)
		if err != nil {
			continue
		}

		statusCode, _, err := response.ToString()
		if err != nil || statusCode != http.StatusOK {
			continue
		}

		return nil
	}

	errMsg := tlog.E(ctx).Msg("startup health probe failed after all retry attempts")

	errx := terror.NewRawTerror(ctx, terror.ErrSvcExecute("http server"), errMsg)

	return errx
}
