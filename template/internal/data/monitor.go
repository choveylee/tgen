package data

// CpuCheckRespData contains the CPU health check result returned by the monitoring endpoint.
type CpuCheckRespData struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// RamCheckRespData contains the memory health check result returned by the monitoring endpoint.
type RamCheckRespData struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}
