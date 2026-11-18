package handler

import (
	"net/http"

	"github.com/choveylee/tlog"
	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/service"
)

// HandleCpuCheck serves the CPU health check endpoint.
func HandleCpuCheck(ginContext *gin.Context) {
	ctx := ginContext.Request.Context()

	cpuCheckRespData, errx := service.CpuCheck(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx).Msgf("Handle cpu check err (cpu check %v)",
			errx)

		SendFailResponse(ginContext, errx.ErrCode(), errMsg)

		return
	}

	statusCode := http.StatusOK
	switch cpuCheckRespData.Status {
	case constant.HealthStatusWarning:
		statusCode = http.StatusTooManyRequests
	case constant.HealthStatusCritical:
		statusCode = http.StatusInternalServerError
	}

	SendResponse(ginContext, statusCode, cpuCheckRespData)
}

// HandleRamCheck serves the memory health check endpoint.
func HandleRamCheck(ginContext *gin.Context) {
	ctx := ginContext.Request.Context()

	ramCheckRespData, errx := service.RamCheck(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx).Msgf("Handle ram check err (ram check %v)",
			errx)

		SendFailResponse(ginContext, errx.ErrCode(), errMsg)

		return
	}

	statusCode := http.StatusOK
	switch ramCheckRespData.Status {
	case constant.HealthStatusWarning:
		statusCode = http.StatusTooManyRequests
	case constant.HealthStatusCritical:
		statusCode = http.StatusInternalServerError
	}

	SendResponse(ginContext, statusCode, ramCheckRespData)
}
