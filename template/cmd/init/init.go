/**
 * @Author: lidonglin
 * @Description:
 * @File:  init.go
 * @Version: 1.0.0
 * @Date: 2023/12/06 09:01
 */

package init

import (
	"context"
	"time"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/tlog"
)

func init() {
	ctx := context.Background()

	timeLocation, err := tcfg.String(tcfg.LocalKey("TIME_LOCATION"))
	if err != nil {
		tlog.E(ctx).Err(err).Msgf("init (%s) err (cfg string %v).",
			"TIME_LOCATION", err)

		return
	}

	location, err := time.LoadLocation(timeLocation)
	if err != nil {
		tlog.E(ctx).Err(err).Msgf("init (%s) err (load location %v).",
			timeLocation, err)

		return
	}

	time.Local = location
}
