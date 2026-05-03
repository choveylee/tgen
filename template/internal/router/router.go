// Package router configures the HTTP routing used by the generated service.
package router

import (
	"context"

	"github.com/choveylee/tcfg"
	"github.com/choveylee/tserver"
	tmiddleware "github.com/choveylee/tserver/middleware"
	"github.com/gin-gonic/gin"
)

// NewRouter constructs the service HTTP router and registers application routes.
func NewRouter(ctx context.Context) *gin.Engine {
	appName := tcfg.DefaultString("APP_NAME", "unknown")

	router := tserver.NewRouter(appName)

	router.Use(tmiddleware.CorsMiddleware())

	// Register monitoring routes.
	registerMonitor(router)

	return router
}
