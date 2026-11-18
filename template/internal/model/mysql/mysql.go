// Package dbmodel configures the MySQL client shared by the service.
package dbmodel

import (
	"context"
	"strings"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/tdb"
	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"
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

	serverDsn := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey(runMode+"::SERVER_MYSQL_DSN"), ""))
	if serverDsn == "" {
		errMsg := tlog.E(ctx).Msg("Init mysql model err (server mysql dsn empty)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server mysql dsn"), errMsg)

		return errx
	}

	maxIdleConns := tcfg.DefaultInt(tcfg.LocalKey("MYSQL_MAX_IDLE_CONNS"), 10)
	if maxIdleConns < 0 {
		errMsg := tlog.E(ctx).Msg("Init mysql model err (mysql max idle connections invalid)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("mysql max idle connections"), errMsg)

		return errx
	}

	maxOpenConns := tcfg.DefaultInt(tcfg.LocalKey("MYSQL_MAX_OPEN_CONNS"), 100)
	if maxOpenConns <= 0 {
		errMsg := tlog.E(ctx).Msg("Init mysql model err (mysql max open connections invalid)")

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
		errMsg := tlog.E(ctx).Err(err).Msg("Init mysql model err (new mysql client)")

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	if client == nil {
		errMsg := tlog.E(ctx).Msg("Init mysql model err (mysql client invalid)")

		errx := terror.NewRawTerror(ctx, terror.ErrSvcAbnormal("mysql client"), errMsg)

		return errx
	}

	err = client.SetMaxIdleConns(maxIdleConns)
	if err != nil {
		_ = client.Close()
		errMsg := tlog.E(ctx).Err(err).Msgf("Init mysql model (max idle connections: %d) err (set max idle connections)",
			maxIdleConns)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	err = client.SetMaxOpenConns(maxOpenConns)
	if err != nil {
		_ = client.Close()
		errMsg := tlog.E(ctx).Err(err).Msgf("Init mysql model (max open connections: %d) err (set max open connections)",
			maxOpenConns)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	serverClient = client

	return nil
}

// Tx begins a transaction that the caller must commit or roll back.
func Tx(ctx context.Context) *gorm.DB {
	return serverClient.Tx(ctx, runMode)
}

// DB returns the request-scoped GORM database handle.
func DB(ctx context.Context) *gorm.DB {
	return serverClient.DB(ctx, runMode)
}
