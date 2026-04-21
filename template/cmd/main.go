/**
 * @Author: lidonglin
 * @Description:
 * @File:  main.go
 * @Version: 1.0.0
 * @Date: 2023/12/06 08:54
 */

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
		tlog.E(ctx).Err(errx).Msgf("main err (run migrate %s).",
			errx)

		return
	}

	// init lib
	errx = lib.InitLib(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msgf("main err (init lib %s).",
			errx)

		return
	}

	// init model
	errx = model.InitModel(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msgf("main err (init model %s).",
			errx)

		return
	}

	// init cron
	errx = crontab.InitCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msgf("main err (init cron %s).",
			errx)

		return
	}

	errx = crontab.StartCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msgf("main err (start cron %s).",
			errx)

		return
	}

	// init service
	errx = service.InitService(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msgf("main err (init service %s).",
			errx)

		return
	}

	httpPort := tcfg.DefaultInt(tcfg.LocalKey("HTTP_PORT"), 8080)

	go func() {
		if err := waitForTcpDial(ctx, httpPort, 30*time.Second); err != nil {
			tlog.W(ctx).Msg("http server tcp not ready within timeout, skip health ping.")

			return
		}

		errx := pingServer(ctx, httpPort)
		if errx != nil {
			tlog.W(ctx).Msg("http server health check failed after tcp ready.")
		} else {
			tlog.I(ctx).Msg("http server deployed success.")
		}
	}()

	tserver.StartHttpServer(ctx, router.NewRouter(ctx), httpPort)
}

func runMigrate(ctx context.Context) *terror.Terror {
	runMode := tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug)

	serverDsn, err := tcfg.String(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_MYSQL_DSN")))
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s::%s) err (cfg string %s).",
			runMode, "SERVER_MYSQL_DSN", err)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	client, err := migrate.New("file://migration", "mysql://"+tutil.MysqlDsnEncode(serverDsn))
	if err != nil {
		serverCfg, err := mysql.ParseDSN(serverDsn)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s) err (parse dsn %s).",
				serverDsn, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		dbName := serverCfg.DBName

		serverCfg.DBName = ""
		tmpDsn := serverCfg.FormatDSN()

		db, err := sql.Open("mysql", tmpDsn)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s, %s) err (open mysql %s).",
				serverDsn, tmpDsn, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		defer db.Close()

		_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + "`" + dbName + "` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s, %s, %s) err (create database %s).",
				serverDsn, tmpDsn, dbName, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		client, err = migrate.New("file://migration", "mysql://"+tutil.MysqlDsnEncode(serverDsn))
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s) err (migrate new %s).",
				serverDsn, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}
	}

	defer func() {
		srcErr, dbErr := client.Close()
		if srcErr != nil || dbErr != nil {
			tlog.W(ctx).Msgf("run migrate close err (source %s, database %s).", srcErr, dbErr)
		}
	}()

	err = client.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s) err (migrate up %s).",
			serverDsn, err)

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

	return fmt.Errorf("timeout waiting for tcp %s", address)
}

// resolvePingBaseUrl 与 HTTP 监听端口对齐：配置为空或为本机回环时，统一使用 httpPort，避免只改 HTTP_PORT 未改 SERVER_PING_HOST。
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

	errMsg := tlog.E(ctx).Msg("ping server err (over limited).")

	errx := terror.NewRawTerror(ctx, terror.ErrSvcExecute("http server"), errMsg)

	return errx
}
