/**
 * @Author: lidonglin
 * @Description:
 * @File:  router.go
 * @Version: 1.0.0
 * @Date: 2023/11/15 21:49
 */

package router

import (
    "context"
 
	"github.com/choveylee/tcfg"
	"github.com/choveylee/tserver"
	"github.com/choveylee/tserver/middleware"
	"github.com/gin-gonic/gin"
 
	"{{domain}}/{{app_name}}/internal/router/middleware"
)
 
func NewRouter(ctx context.Context) *gin.Engine {
	appName := tcfg.DefaultString("APP_NAME", "unknown")
 
	router := tserver.NewRouter(appName)
 
	router.Use(tmiddleware.CorsMiddleware())

	// register monitor
	registerMonitor(router)

	return router
}