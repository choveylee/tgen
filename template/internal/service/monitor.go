/**
 * @Author: lidonglin
 * @Description:
 * @File:  monitor.go
 * @Version: 1.0.0
 * @Date: 2023/06/27 15:14
 */

package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/choveylee/terror"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"

    "{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/data"
)

func CpuCheck(ctx context.Context) (*data.CpuCheckRespData, *terror.Terror) {
	cores, _ := cpu.Counts(false)

	avgStat, _ := load.Avg()
	load1 := avgStat.Load1
	load5 := avgStat.Load5
	load15 := avgStat.Load15

	cpuCheckRespData := &data.CpuCheckRespData{
		StatusCode: http.StatusOK,

		Status: "ErrorCodeOK",
	}

	if load5 >= float64(cores-1) {
		cpuCheckRespData.StatusCode = http.StatusInternalServerError
		cpuCheckRespData.Status = "CRITICAL"
	} else if load5 >= float64(cores-2) {
		cpuCheckRespData.StatusCode = http.StatusTooManyRequests
		cpuCheckRespData.Status = "WARNING"
	}

	cpuCheckRespData.Detail = fmt.Sprintf("%s - Load average: %.2f, %.2f, %.2f | Cores: %d", cpuCheckRespData.Status, load1, load5, load15, cores)

	return cpuCheckRespData, nil
}

func RamCheck(ctx context.Context) (*data.RamCheckRespData, *terror.Terror) {
	virtualMemoryStat, _ := mem.VirtualMemory()

	usedMB := int(virtualMemoryStat.Used) / constant.MB
	usedGB := int(virtualMemoryStat.Used) / constant.GB

	totalMB := int(virtualMemoryStat.Total) / constant.MB
	totalGB := int(virtualMemoryStat.Total) / constant.GB

	usedPercent := int(virtualMemoryStat.UsedPercent)

	ramCheckRespData := &data.RamCheckRespData{
		StatusCode: http.StatusOK,

		Status: "ErrorCodeOK",
	}

	if usedPercent >= 95 {
		ramCheckRespData.StatusCode = http.StatusInternalServerError
		ramCheckRespData.Status = "CRITICAL"
	} else if usedPercent >= 90 {
		ramCheckRespData.StatusCode = http.StatusTooManyRequests
		ramCheckRespData.Status = "WARNING"
	}

	ramCheckRespData.Detail = fmt.Sprintf("%s - Free space: %dMB (%dGB) / %dMB (%dGB) | Used: %d%%", ramCheckRespData.Status, usedMB, usedGB, totalMB, totalGB, usedPercent)

	return ramCheckRespData, nil
}
