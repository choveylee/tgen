/**
 * @Author: lidonglin
 * @Description:
 * @File:  monitor.go
 * @Version: 1.0.0
 * @Date: 2023/06/27 14:50
 */

package handler

import (
	"net/http"

	"github.com/choveylee/tlog"
	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/data"
	"{{domain}}/{{app_name}}/internal/service"
)

func HandleHealthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func HandleCpuCheck(c *gin.Context) {
	ctx := c.Request.Context()

	cpuCheckRespData, errx := service.CpuCheck(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msgf("handle cpu check err (cpu check %s).", errx)
		SendFailResponse(c, errx.ErrCode(), errx.Error())

		return
	}

	c.JSON(cpuCheckRespData.StatusCode, data.Response{
		Code: constant.ErrorCodeOK,
		Data: cpuCheckRespData,
	})
}

func HandleRamCheck(c *gin.Context) {
	ctx := c.Request.Context()

	ramCheckRespData, errx := service.RamCheck(ctx)
	if errx != nil {
		tlog.E(ctx).Err(errx).Msgf("handle ram check err (ram check %s).", errx)
		SendFailResponse(c, errx.ErrCode(), errx.Error())

		return
	}

	c.JSON(ramCheckRespData.StatusCode, data.Response{
		Code: constant.ErrorCodeOK,
		Data: ramCheckRespData,
	})
}
