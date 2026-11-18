// Command orm-gen generates GORM query code for the service schema.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/tdb"
	"github.com/choveylee/tlog"
	"gorm.io/gen"

	"{{domain}}/{{app_name}}/internal/const"
)

func main() {
	ctx := context.Background()
	if err := generateQueries(ctx); err != nil {
		tlog.E(ctx).Err(err).Msg("Generate ORM queries err")

		os.Exit(1)
	}
}

// generateQueries builds the sample query API from the migrated template table.
func generateQueries(ctx context.Context) error {
	runMode := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug))
	serverDSN := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("SERVER_MYSQL_DSN"), ""))
	if serverDSN == "" {
		return errors.New("server mysql dsn must be configured")
	}

	serverClient, err := tdb.NewMysqlClient(ctx, serverDSN)
	if err != nil {
		return fmt.Errorf("open the configured MySQL database (error type %T)", err)
	}

	defer serverClient.Close()

	database := serverClient.DB(ctx, runMode)
	if !database.Migrator().HasTable("template") {
		return errors.New("template table must exist before ORM generation")
	}

	generator := gen.NewGenerator(gen.Config{
		OutPath: "./model/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})
	generator.UseDB(database)
	model := generator.GenerateModel("template")
	generator.ApplyBasic(model)
	generator.ApplyInterface(func(TemplateMethod) {}, model)
	generator.Execute()

	return nil
}
