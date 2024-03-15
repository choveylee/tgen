/**
 * @Author: lidonglin
 * @Description:
 * @File:  cors.go
 * @Version: 1.0.0
 * @Date: 2023/11/16 21:14
 */

package middleware

import (
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
	    c.Next()
 	}
}
 