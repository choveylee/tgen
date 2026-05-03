// Package crontab loads configuration for scheduled jobs and registers them with the cron runtime.
package crontab

import (
	"context"
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
	testSyncCron = tcfg.DefaultString(tcfg.LocalKey("TEST_SYNC_CRON"), "")

	return nil
}

// StartCron registers the configured cron jobs.
func StartCron(ctx context.Context) *terror.Terror {
	cronRedisClient := redmodel.GetCronRedisClient()

	if testSyncCron != "" {
		_, err := tcron.RegisterSingletonCron(testSyncCron, runTestSync, cronRedisClient, 10*time.Minute)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("cron job registration failed for schedule %q",
				testSyncCron)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}
	}

	return nil
}
