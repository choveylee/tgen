/**
 * @Author: lidonglin
 * @Description:
 * @File:  crontab.go
 * @Version: 1.0.0
 * @Date: 2023/12/06 09:22
 */

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

func InitCron(ctx context.Context) *terror.Terror {
	testSyncCron = tcfg.DefaultString(tcfg.LocalKey("TEST_SYNC_CRON"), "")

	return nil
}

func StartCron(ctx context.Context) *terror.Terror {
	cronRedisClient := redmodel.GetCronRedisClient()

	if testSyncCron != "" {
		_, err := tcron.RegisterSingletonCron(testSyncCron, runTestSync, cronRedisClient, 10*time.Minute)
		if err != nil {
			errMsg := tlog.E(ctx).Err(err).Msgf("start cron (%s) err (register test sync %s).",
				testSyncCron, err)

			errx := terror.NewRawTerror(ctx, err, errMsg)

			return errx
		}
	}

	return nil
}
