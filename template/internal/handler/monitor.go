package handler

import (
	"github.com/choveylee/tlog"
	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/service"
)

// HandleCpuCheck serves the CPU health check endpoint.
func HandleCpuCheck(c *gin.Context) {
	ctx := c.Request.Context()

	cpuCheckRespData, errx := service.CpuCheck(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx).Msgf("Handle cpu check err (cpu check %v)",
			errx)

		SendFailResponse(c, errx.ErrCode(), errMsg)

		return
	}

	SendPassResponse(c, cpuCheckRespData)
}

// HandleRamCheck serves the memory health check endpoint.
func HandleRamCheck(c *gin.Context) {
	ctx := c.Request.Context()

	ramCheckRespData, errx := service.RamCheck(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx).Msgf("Handle ram check err (ram check %v)",
			errx)

		SendFailResponse(c, errx.ErrCode(), errMsg)

		return
	}

	SendPassResponse(c, ramCheckRespData)
}
