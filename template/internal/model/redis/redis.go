// Package redmodel configures the Redis client shared by the service.
package redmodel

import (
	"context"
	"strings"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/tdb"
	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"
	"github.com/redis/go-redis/v9"

	"{{domain}}/{{app_name}}/internal/const"
)

var (
	runMode string

	serverClient *tdb.RedisClient
)

// GetCronRedisClient returns the Redis client used for cron coordination.
func GetCronRedisClient() *redis.Client {
	return serverClient.Client()
}

// InitRedisModel initializes the shared Redis client.
func InitRedisModel(ctx context.Context) *terror.Terror {
	runMode = strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug))

	serverAddress := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey(runMode+"::SERVER_REDIS_ADDRESS"), ""))
	if serverAddress == "" {
		errMsg := tlog.E(ctx).Msg("Init redis model err (server redis address empty)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server redis address"), errMsg)

		return errx
	}

	serverPassword := tcfg.DefaultString(tcfg.LocalKey(runMode+"::SERVER_REDIS_PASSWORD"), "")

	serverPoolSize := tcfg.DefaultInt(tcfg.LocalKey("SERVER_REDIS_POOLSIZE"), 100)
	if serverPoolSize <= 0 {
		errMsg := tlog.E(ctx).Msg("Init redis model err (server redis pool size invalid)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server redis pool size"), errMsg)

		return errx
	}

	client, err := tdb.NewRedisClient(ctx, serverAddress, serverPassword, serverPoolSize)
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("Init redis model (server address: %q, pool size: %d) err (new redis client)",
			serverAddress, serverPoolSize)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	if client == nil {
		errMsg := tlog.E(ctx).Msg("Init redis model err (redis client invalid)")

		errx := terror.NewRawTerror(ctx, terror.ErrSvcAbnormal("redis client"), errMsg)

		return errx
	}

	serverClient = client

	return nil
}
