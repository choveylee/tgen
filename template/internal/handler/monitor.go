package handler

import (
	"github.com/choveylee/tlog"
	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/data"
	"{{domain}}/{{app_name}}/internal/service"
)

// HandleCpuCheck serves the CPU health check endpoint.
func HandleCpuCheck(c *gin.Context) {
	ctx := c.Request.Context()

	cpuCheckRespData, errx := service.CpuCheck(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("CPU health check request failed")
		SendFailResponse(c, errx.ErrCode(), errx.Error())

		return
	}

	c.JSON(cpuCheckRespData.StatusCode, data.Response{
		Code: constant.ErrorCodeOK,
		Data: cpuCheckRespData,
	})
}

// HandleRamCheck serves the memory health check endpoint.
func HandleRamCheck(c *gin.Context) {
	ctx := c.Request.Context()

	ramCheckRespData, errx := service.RamCheck(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msg("Memory health check request failed")
		SendFailResponse(c, errx.ErrCode(), errx.Error())

		return
	}

	c.JSON(ramCheckRespData.StatusCode, data.Response{
		Code: constant.ErrorCodeOK,
		Data: ramCheckRespData,
	})
}
