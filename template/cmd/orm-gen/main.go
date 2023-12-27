/**
 * @Author: lidonglin
 * @Description:
 * @File:  main.go
 * @Version: 1.0.0
 * @Date: 2023/12/26 14:21
 */

package main

import (
	"context"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/tdb"
	"github.com/choveylee/tlog"
	"gorm.io/gen"
)

func main() {
	ctx := context.Background()

	runMode := tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), "debug")

	serverDsn, err := tcfg.String(tcfg.LocalKey("SERVER_MYSQL_DSN"))
	if err != nil {
		tlog.E(ctx).Err(err).Msgf("main (%s) err (cfg string %v).",
			"SERVER_MYSQL_DSN", err)

		return
	}

	serverClient, err := tdb.NewMysqlClient(ctx, serverDsn)
	if err != nil {
		tlog.E(ctx).Err(err).Msgf("main (%s) err (new mysql client %v).",
			serverDsn, err)

		return
	}

	generator := gen.NewGenerator(gen.Config{
		OutPath: "./model/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})

	sqlDB := serverClient.DB(ctx, runMode)

	generator.UseDB(sqlDB)

	// Generate basic type-safe DAO API for table `template` following conventions
	generator.ApplyBasic(generator.GenerateModel("template"))

	// Generate Type Safe API with Dynamic SQL defined on Query interface for `template`
	generator.ApplyInterface(func(TemplateMethod) {}, generator.GenerateModel("template"))

	// Generate the code
	generator.Execute()
}
