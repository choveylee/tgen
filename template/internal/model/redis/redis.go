// Package redmodel configures the Redis client shared by the service.
package redmodel

import (
	"context"
	"fmt"
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

	serverAddress := strings.TrimSpace(tcfg.DefaultString(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_REDIS_ADDRESS")), ""))
	if serverAddress == "" {
		errMsg := tlog.E(ctx).Msg("Redis initialization failed because the server address is not configured")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server redis address"), errMsg)

		return errx
	}

	serverPassword := tcfg.DefaultString(tcfg.LocalKey(runMode+"::SERVER_REDIS_PASSWORD"), "")

	serverPoolSize := tcfg.DefaultInt(tcfg.LocalKey("SERVER_REDIS_POOLSIZE"), 100)
	if serverPoolSize <= 0 {
		errMsg := tlog.E(ctx).Msg("Initialize model err (server redis pool size invalid)")

		errx := terror.NewRawTerror(ctx, terror.ErrConfInvalid("server redis pool size"), errMsg)

		return errx
	}

	var err error

	serverClient, err = tdb.NewRedisClient(ctx, serverAddress, serverPassword, serverPoolSize)
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("Redis initialization failed while creating the client for address %q with pool size %d",
			serverAddress, serverPoolSize)

		errx := terror.NewRawTerror(ctx, err, errMsg)

		return errx
	}

	return nil
}
