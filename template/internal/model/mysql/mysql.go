/**
 * @Author: lidonglin
 * @Description:
 * @File:  mysql.go
 * @Version: 1.0.0
 * @Date: 2023/11/15 21:22
 */

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
 
	 "{{domain}}/{{app_name}}/internal/const"
 )
 
 var (
	 runMode string
 
	 serverClient *tdb.MysqlClient
 )
 
 func InitMysqlModel(ctx context.Context) *terror.Terror {
	 runMode = tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeRelease)
 
	 serverDsn, err := tcfg.String(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_MYSQL_DSN")))
	 if err != nil {
		 errMsg := tlog.E(ctx).Err(err).Msgf("init mysql model (%s::%s) err (cfg string %v).",
			 runMode, "SERVER_MYSQL_DSN", err)
 
		 errx := terror.NewRawTerror(ctx, err, errMsg)
 
		 return errx
	 }
 
	 if runMode == constant.RunModeDebug {
		 serverClient, err = tdb.NewMysqlClientWithLog(ctx, serverDsn)
	 } else {
		 serverClient, err = tdb.NewMysqlClient(ctx, serverDsn)
	 }
	 if err != nil {
		 errMsg := tlog.E(ctx).Err(err).Msgf("init mysql model (%s) err (new mysql client %v).",
			 serverDsn, err)
 
		 errx := terror.NewRawTerror(ctx, err, errMsg)
 
		 return errx
	 }
 
	 maxIdleConns := tcfg.DefaultInt(tcfg.LocalKey("MYSQL_MAX_IDLE_CONNS"), 10)
 
	 err = serverClient.SetMaxIdleConns(maxIdleConns)
	 if err != nil {
		 errMsg := tlog.E(ctx).Err(err).Msgf("init mysql model (%d) err (set max idle conns %v).",
			 maxIdleConns, err)
 
		 errx := terror.NewRawTerror(ctx, err, errMsg)
 
		 return errx
	 }
 
	 maxOpenConns := tcfg.DefaultInt(tcfg.LocalKey("MYSQL_MAX_OPEN_CONNS"), 100)
 
	 err = serverClient.SetMaxOpenConns(maxOpenConns)
	 if err != nil {
		 errMsg := tlog.E(ctx).Err(err).Msgf("init mysql model (%d) err (set max open conns %v).",
			 maxOpenConns, err)
 
		 errx := terror.NewRawTerror(ctx, err, errMsg)
 
		 return errx
	 }
 
	 return nil
 }
 