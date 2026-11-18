// Package constant defines shared constants used across the service.
package constant

const (
	RunModeDebug   = "debug"
	RunModeRelease = "release"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

const (
	SortSeqAsc  = "asc"
	SortSeqDesc = "des"
)

// Health states are shared by the monitor service and its HTTP handlers.
const (
	HealthStatusOK       = "OK"
	HealthStatusWarning  = "WARNING"
	HealthStatusCritical = "CRITICAL"
)
