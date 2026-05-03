// Package middleware provides HTTP middleware used by the service router.
package middleware

import (
	"github.com/gin-gonic/gin"
)

// AuthMiddleware returns a placeholder authentication middleware.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
