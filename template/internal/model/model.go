/**
 * @Author: lidonglin
 * @Description:
 * @File:  model.go
 * @Version: 1.0.0
 * @Date: 2023/11/15 18:10
 */

package model

import (
	"context"
 
	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"
 
	"{{domain}}/{{app_name}}/internal/model/mysql"
	"{{domain}}/{{app_name}}/internal/model/redis"
)
 
func InitModel(ctx context.Context) *terror.Terror {
	errx := dbmodel.InitMysqlModel(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx.Error()).Msgf("init model err (init mysql model %v).", errx.Error())
 		errx.AttachErrMsg(errMsg)
 
		return errx
	}
 
	errx = redmodel.InitRedisModel(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx.Error()).Msgf("init model err (init redis model %v).", errx.Error())
 		errx.AttachErrMsg(errMsg)
 
		return errx
	}
 
	return nil
}
 