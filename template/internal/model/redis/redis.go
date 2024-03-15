/**
 * @Author: lidonglin
 * @Description:
 * @File:  redis.go
 * @Version: 1.0.0
 * @Date: 2023/11/15 21:23
 */

package redmodel

import (
	"context"
	"fmt"
 
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
 
func GetCronRedisClient(ctx context.Context) *redis.Client {
	return serverClient.Client(ctx)
}
 
func InitRedisModel(ctx context.Context) *terror.Terror {
	runMode = tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeRelease)
 
	serverAddress := tcfg.DefaultString(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_REDIS_ADDRESS")), "")
	if serverAddress == "" {
		errMsg := tlog.E(ctx).Msgf("init redis model err (server redis address illegal).")
 
		errx := terror.NewRawTerror(ctx, terror.ErrConfIllegal("server redis address"), errMsg)
 
		return errx
	}
 
	serverPassword := tcfg.DefaultString(fmt.Sprintf("%s::%s", runMode, tcfg.LocalKey("SERVER_REDIS_PASSWORD")), "")
	 
	serverPoolSize := tcfg.DefaultInt(tcfg.LocalKey("SERVER_REDIS_POOLSIZE"), 100)

    var err error

	serverClient, err = tdb.NewRedisClient(ctx, serverAddress, serverPassword, serverPoolSize)
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("init redis model (%s, %s, %d) err (new redis client %v).",
			serverAddress, serverPassword, serverPoolSize, err)
 
		errx := terror.NewRawTerror(ctx, err, errMsg)
 
		return errx
	}
 
	return nil
}
 