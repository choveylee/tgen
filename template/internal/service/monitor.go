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

func CpuCheck(ctx context.Context) (*data.CpuCheckData, *terror.Terror) {
	cores, _ := cpu.Counts(false)

	avgStat, _ := load.Avg()
	load1 := avgStat.Load1
	load5 := avgStat.Load5
	load15 := avgStat.Load15

	cpuCheckData := &data.CpuCheckData{
		StatusCode: http.StatusOK,

		Status: "ErrorCodeOK",
	}

	if load5 >= float64(cores-1) {
		cpuCheckData.StatusCode = http.StatusInternalServerError
		cpuCheckData.Status = "CRITICAL"
	} else if load5 >= float64(cores-2) {
		cpuCheckData.StatusCode = http.StatusTooManyRequests
		cpuCheckData.Status = "WARNING"
	}

	cpuCheckData.Detail = fmt.Sprintf("%s - Load average: %.2f, %.2f, %.2f | Cores: %d", cpuCheckData.Status, load1, load5, load15, cores)

	return cpuCheckData, nil
}

func RamCheck(ctx context.Context) (*data.RamCheckData, *terror.Terror) {
	virtualMemoryStat, _ := mem.VirtualMemory()

	usedMB := int(virtualMemoryStat.Used) / constant.MB
	usedGB := int(virtualMemoryStat.Used) / constant.GB

	totalMB := int(virtualMemoryStat.Total) / constant.MB
	totalGB := int(virtualMemoryStat.Total) / constant.GB

	usedPercent := int(virtualMemoryStat.UsedPercent)

	ramCheckData := &data.RamCheckData{
		StatusCode: http.StatusOK,

		Status: "ErrorCodeOK",
	}

	if usedPercent >= 95 {
		ramCheckData.StatusCode = http.StatusInternalServerError
		ramCheckData.Status = "CRITICAL"
	} else if usedPercent >= 90 {
		ramCheckData.StatusCode = http.StatusTooManyRequests
		ramCheckData.Status = "WARNING"
	}

	ramCheckData.Detail = fmt.Sprintf("%s - Free space: %dMB (%dGB) / %dMB (%dGB) | Used: %d%%", ramCheckData.Status, usedMB, usedGB, totalMB, totalGB, usedPercent)

	return ramCheckData, nil
}
