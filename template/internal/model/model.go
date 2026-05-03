// Package model initializes storage clients and repository-layer dependencies.
package model

import (
	"context"

	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"

	"{{domain}}/{{app_name}}/internal/model/mysql"
	"{{domain}}/{{app_name}}/internal/model/redis"
)

// InitModel initializes all persistence-layer dependencies.
func InitModel(ctx context.Context) *terror.Terror {
	errx := dbmodel.InitMysqlModel(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx).Msg("model initialization failed while setting up MySQL dependencies")
		errx.AttachErrMsg(errMsg)

		return errx
	}

	errx = redmodel.InitRedisModel(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx).Msg("model initialization failed while setting up Redis dependencies")
		errx.AttachErrMsg(errMsg)

		return errx
	}

	return nil
}
