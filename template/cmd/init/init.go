// Package init performs process-level initialization required before the service starts.
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
		tlog.F(ctx).Err(err).Msgf("process initialization failed while reading configuration key %s", "TIME_LOCATION")
	}

	location, err := time.LoadLocation(timeLocation)
	if err != nil {
		tlog.F(ctx).Err(err).Msgf("process initialization failed while loading time location %q", timeLocation)
	}

	time.Local = location
}
