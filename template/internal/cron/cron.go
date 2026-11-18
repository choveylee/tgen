// Package crontab loads configuration for scheduled jobs and registers them with the cron runtime.
package crontab

import (
	"context"
	"strings"
	"time"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/tcron"
	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"

	redmodel "{{domain}}/{{app_name}}/internal/model/redis"
)

var (
	testSyncCron string
)

// InitCron loads cron configuration.
func InitCron(ctx context.Context) *terror.Terror {
	testSyncCron = strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("TEST_SYNC_CRON"), ""))

	return nil
}

// StartCron registers the configured cron jobs.
func StartCron(ctx context.Context) *terror.Terror {
	cronRedisClient := redmodel.GetCronRedisClient()

	if testSyncCron != "" {
		_, err := tcron.RegisterSingletonCron(testSyncCron, runTestSync, cronRedisClient, 10*time.Minute)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("Start cron (test sync cron: %q) err (register singleton cron %v)",
				testSyncCron, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}
	}

	return nil
}
