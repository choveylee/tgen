// Package dbmodel configures the MySQL client shared by the service.
package dbmodel

import (
	"context"
	"fmt"
	"strings"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/tdb"
	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"

	"{{domain}}/{{app_name}}/internal/const"
)

var (
	runMode string

	serverClient *tdb.MysqlClient
)

// InitMysqlModel initializes the shared MySQL client.
func InitMysqlModel(ctx context.Context) *terror.Terror {
	runMode = strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug))

	serverDsn := strings.TrimSpace(tcfg.DefaultString(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_MYSQL_DSN")), ""))
	if serverDsn == "" {
		errMsg := tlog.E(ctx).Msg("MySQL initialization failed because the server DSN is not configured")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server mysql dsn"), errMsg)

		return errx
	}

	maxIdleConns := tcfg.DefaultInt(tcfg.LocalKey("MYSQL_MAX_IDLE_CONNS"), 10)
	if maxIdleConns < 0 {
		errMsg := tlog.E(ctx).Msg("Initialize model err (mysql max idle connections invalid)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("mysql max idle connections"), errMsg)

		return errx
	}

	maxOpenConns := tcfg.DefaultInt(tcfg.LocalKey("MYSQL_MAX_OPEN_CONNS"), 100)
	if maxOpenConns <= 0 {
		errMsg := tlog.E(ctx).Msg("Initialize model err (mysql max open connections invalid)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("mysql max open connections"), errMsg)

		return errx
	}

	var (
		client *tdb.MysqlClient
		err    error
	)

	if runMode == constant.RunModeDebug {
		client, err = tdb.NewMysqlClientWithLog(ctx, serverDsn)
	} else {
		client, err = tdb.NewMysqlClient(ctx, serverDsn)
	}

	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msg("MySQL initialization failed while creating the client")

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	err = client.SetMaxIdleConns(maxIdleConns)
	if err != nil {
		_ = client.Close()
		errMsg := tlog.E(ctx).Err(err).Msgf("MySQL initialization failed while setting the maximum idle connection count to %d",
			maxIdleConns)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	err = client.SetMaxOpenConns(maxOpenConns)
	if err != nil {
		_ = client.Close()
		errMsg := tlog.E(ctx).Err(err).Msgf("MySQL initialization failed while setting the maximum open connection count to %d",
			maxOpenConns)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	serverClient = client

	return nil
}

// Tx returns the request-scoped GORM database handle.
func Tx(ctx context.Context) *gorm.DB {
	return serverClient.Tx(ctx, runMode)
}

// DB returns the request-scoped GORM database handle.
func DB(ctx context.Context) *gorm.DB {
	return serverClient.DB(ctx, runMode)
}
