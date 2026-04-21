/**
 * @Author: lidonglin
 * @Description:
 * @File:  monitor.go
 * @Version: 1.0.0
 * @Date: 2023/06/27 15:08
 */

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/handler"
)

func registerMonitor(router *gin.Engine) {
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	router.GET("/cpu-check", handler.HandleCpuCheck)
	router.GET("/ram-check", handler.HandleRamCheck)
}
