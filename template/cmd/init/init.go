// Package init performs process-level initialization required before the service starts.
package init

import (
	"context"
	"strings"
	"time"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/terror"
	"github.com/choveylee/tlog"

	"{{domain}}/{{app_name}}/internal/const"
)

// init validates the run mode and applies the process timezone before startup.
func init() {
	ctx := context.Background()

	runMode := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), constant.RunModeDebug))
	if runMode != constant.RunModeDebug && runMode != constant.RunModeRelease {
		tlog.F(ctx).Err(terror.ErrConfInvalid("run mode")).Msgf("Initialize process (run mode: %q) err (run mode must be debug or release)",
			runMode)
	}

	timeLocation := strings.TrimSpace(tcfg.DefaultString(tcfg.LocalKey("TIME_LOCATION"), "Asia/Shanghai"))
	if timeLocation == "" {
		tlog.F(ctx).Err(terror.ErrConfInvalid("time location")).Msg("Initialize process err (time location empty)")
	}

	location, err := time.LoadLocation(timeLocation)
	if err != nil {
		tlog.F(ctx).Err(err).Msgf("Initialize process (time location: %q) err (load location %v)",
			timeLocation, err)
	}

	time.Local = location
}
