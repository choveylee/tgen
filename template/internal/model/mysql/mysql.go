// Package dbmodel configures the MySQL client shared by the service.
package dbmodel

import (
	"context"
	"fmt"

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
	runMode = tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug)

	serverDsn := tcfg.DefaultString(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_MYSQL_DSN")), "")
	if serverDsn == "" {
		errMsg := tlog.E(ctx).Msg("MySQL initialization failed because the server DSN is not configured")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("SERVER_MYSQL_DSN"), errMsg)

		return errx
	}

	var err error

	if runMode == constant.RunModeDebug {
		serverClient, err = tdb.NewMysqlClientWithLog(ctx, serverDsn)
	} else {
		serverClient, err = tdb.NewMysqlClient(ctx, serverDsn)
	}
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msg("MySQL initialization failed while creating the client")

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	maxIdleConns := tcfg.DefaultInt(tcfg.LocalKey("MYSQL_MAX_IDLE_CONNS"), 10)

	err = serverClient.SetMaxIdleConns(maxIdleConns)
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("MySQL initialization failed while setting the maximum idle connection count to %d",
			maxIdleConns)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	maxOpenConns := tcfg.DefaultInt(tcfg.LocalKey("MYSQL_MAX_OPEN_CONNS"), 100)

	err = serverClient.SetMaxOpenConns(maxOpenConns)
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("MySQL initialization failed while setting the maximum open connection count to %d",
			maxOpenConns)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	return nil
}

// Tx returns the request-scoped GORM database handle.
func Tx(ctx context.Context) *gorm.DB {
	return serverClient.Tx(ctx, runMode)
}
