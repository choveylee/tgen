/**
 * @Author: lidonglin
 * @Description:
 * @File:  monitor.go
 * @Version: 1.0.0
 * @Date: 2023/06/27 14:50
 */

package handler

import (
	"github.com/choveylee/tlog"
	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/service"
)

func HandleCpuCheck(c *gin.Context) {
	ctx := c.Request.Context()

	cpuCheckRespData, errx := service.CpuCheck(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx.Error()).Msgf("handle cpu check err (cpu check %v).",
			errx.Error())

		SendFailResponse(c, errx.ErrCode(), errMsg)

		return
	}

	SendPassResponse(c, cpuCheckRespData)
}

func HandleRamCheck(c *gin.Context) {
	ctx := c.Request.Context()

	ramCheckRespData, errx := service.RamCheck(ctx)
	if errx != nil {
		errMsg := tlog.E(ctx).Err(errx.Error()).Msgf("handle ram check err (ram check %v).",
			errx.Error())

		SendFailResponse(c, errx.ErrCode(), errMsg)

		return
	}

	SendPassResponse(c, ramCheckRespData)
}
