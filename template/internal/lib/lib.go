// Package lib initializes shared library integrations required by the service.
package lib

import (
	"context"

	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"
)

// InitLib initializes shared library dependencies required by the service.
func InitLib(ctx context.Context) *terror.Terror {
	errx := initServerPing(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx).Msg("Init lib err (init server ping)")
		errx.AttachErrMsg(errMsg)

		return errx
	}

	return nil
}
