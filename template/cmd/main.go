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
	"net/http"
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
		tlog.E(ctx).Err(errx.Error()).Msgf("main err (run migrate %v).",
			errx.Error())

		return
	}

	// init lib
	errx = lib.InitLib(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx.Error()).Msgf("main err (init lib %v).",
			errx.Error())

		return
	}

	// init model
	errx = model.InitModel(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx.Error()).Msgf("main err (init model %v).",
			errx.Error())

		return
	}

	// init cron
	errx = crontab.InitCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx.Error()).Msgf("main err (init cron %v).",
			errx.Error())

		return
	}

	errx = crontab.StartCron(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx.Error()).Msgf("main err (start cron %v).",
			errx.Error())

		return
	}

	// init service
	errx = service.InitService(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx.Error()).Msgf("main err (init service %v).",
			errx.Error())

		return
	}

	// start ping server
	go func() {
		errx := pingServer(ctx)
		if errx != nil {
			tlog.W(ctx).Msg("http server not ready, may took too long to start up.")
		} else {
			tlog.I(ctx).Msg("http server deployed success.")
		}
	}()

	// init http server
	httpPort := tcfg.DefaultInt(tcfg.LocalKey("HTTP_PORT"), 80)

	tserver.StartHttpServer(ctx, router.NewRouter(ctx), httpPort)
}

func runMigrate(ctx context.Context) *terror.Terror {
	runMode := tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), "debug")

	serverDsn, err := tcfg.String(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_MYSQL_DSN")))
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s::%s) err (cfg string %v).",
			runMode, "SERVER_MYSQL_DSN", err)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	client, err := migrate.New("file://migration", "mysql://"+tutil.DsnEncode(serverDsn))
	if err != nil {
		serverCfg, err := mysql.ParseDSN(serverDsn)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s) err (parse dsn %v).",
				serverDsn, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		dbName := serverCfg.DBName

		serverCfg.DBName = ""
		tmpDsn := serverCfg.FormatDSN()

		db, err := sql.Open("mysql", tmpDsn)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s, %s) err (open mysql %v).",
				serverDsn, tmpDsn, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		defer db.Close()

		_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + "`" + dbName + "` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s, %s, %s) err (create database %v).",
				serverDsn, tmpDsn, dbName, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}

		client, err = migrate.New("file://migration", "mysql://"+tutil.DsnEncode(serverDsn))
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s) err (migrate new %v).",
				serverDsn, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}
	}

	err = client.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		errMsg := tlog.E(ctx).Err(err).Msgf("run migrate (%s) err (migrate up %v).",
			serverDsn, err)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	return nil
}

func pingServer(ctx context.Context) *terror.Terror {
	pingCount := tcfg.DefaultInt(tcfg.LocalKey("SERVER_PING_COUNT"), 3)
	pingHost := tcfg.DefaultString(tcfg.LocalKey("SERVER_PING_HOST"), "http://127.0.0.1")

	for i := 0; i < pingCount; i++ {
		time.Sleep(time.Second)

		pingUrl := fmt.Sprintf("%s%s", pingHost, "/healthz")

		// ping the server by sending a GET request to `/healthz`.
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
