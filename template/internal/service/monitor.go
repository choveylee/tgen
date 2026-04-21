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
	"github.com/choveylee/tlog"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"

	"{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/data"
)

const (
	loadPerCoreCritical = 1.0
	loadPerCoreWarning  = 0.85
)

func CpuCheck(ctx context.Context) (*data.CpuCheckRespData, *terror.Terror) {
	cores, err := cpu.Counts(false)
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("cpu check err (cpu counts %s).", err)

		return nil, terror.NewRawTerror(ctx, err, errMsg)
	}

	if cores < 1 {
		cores = 1
	}

	avgStat, err := load.Avg()
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("cpu check err (load avg %s).", err)

		return nil, terror.NewRawTerror(ctx, err, errMsg)
	}

	load1 := avgStat.Load1
	load5 := avgStat.Load5
	load15 := avgStat.Load15

	loadPerCore := load5 / float64(cores)

	cpuCheckRespData := &data.CpuCheckRespData{
		StatusCode: http.StatusOK,

		Status: "OK",
	}

	if loadPerCore >= loadPerCoreCritical {
		cpuCheckRespData.StatusCode = http.StatusInternalServerError
		cpuCheckRespData.Status = "CRITICAL"
	} else if loadPerCore >= loadPerCoreWarning {
		cpuCheckRespData.StatusCode = http.StatusTooManyRequests
		cpuCheckRespData.Status = "WARNING"
	}

	cpuCheckRespData.Detail = fmt.Sprintf("%s - Load average: %.2f, %.2f, %.2f | Load/core: %.2f | Cores: %d",
		cpuCheckRespData.Status, load1, load5, load15, loadPerCore, cores,
	)

	return cpuCheckRespData, nil
}

func RamCheck(ctx context.Context) (*data.RamCheckRespData, *terror.Terror) {
	virtualMemoryStat, err := mem.VirtualMemory()
	if err != nil {
		errMsg := tlog.E(ctx).Err(err).Msgf("ram check err (virtual memory %s).", err)

		return nil, terror.NewRawTerror(ctx, err, errMsg)
	}

	usedMB := int(virtualMemoryStat.Used) / constant.MB
	usedGB := int(virtualMemoryStat.Used) / constant.GB

	totalMB := int(virtualMemoryStat.Total) / constant.MB
	totalGB := int(virtualMemoryStat.Total) / constant.GB

	usedPercent := int(virtualMemoryStat.UsedPercent)

	ramCheckRespData := &data.RamCheckRespData{
		StatusCode: http.StatusOK,

		Status: "OK",
	}

	if usedPercent >= 95 {
		ramCheckRespData.StatusCode = http.StatusInternalServerError
		ramCheckRespData.Status = "CRITICAL"
	} else if usedPercent >= 90 {
		ramCheckRespData.StatusCode = http.StatusTooManyRequests
		ramCheckRespData.Status = "WARNING"
	}

	ramCheckRespData.Detail = fmt.Sprintf("%s - Used: %dMB (%dGB) / Total: %dMB (%dGB) | Used: %d%%",
		ramCheckRespData.Status, usedMB, usedGB, totalMB, totalGB, usedPercent)

	return ramCheckRespData, nil
}
