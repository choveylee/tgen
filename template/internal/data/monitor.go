package data

// CpuCheckRespData contains the CPU health check result returned by the monitoring endpoint.
type CpuCheckRespData struct {
	StatusCode int `json:"-"`

	Status string `json:"status"`
	Detail string `json:"detail"`
}

// RamCheckRespData contains the memory health check result returned by the monitoring endpoint.
type RamCheckRespData struct {
	StatusCode int `json:"-"`

	Status string `json:"status"`
	Detail string `json:"detail"`
}
