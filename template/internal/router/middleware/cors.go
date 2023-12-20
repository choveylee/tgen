/**
 * @Author: lidonglin
 * @Description:
 * @File:  cors.go
 * @Version: 1.0.0
 * @Date: 2023/11/16 21:14
 */

 package middleware

 import (
	 "net/http"
 
	 "github.com/gin-gonic/gin"
 )
 
 func Cors() gin.HandlerFunc {
	 return func(c *gin.Context) {
		 c.Header("Access-Control-Allow-Origin", "*")
		 c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token, session, X_Requested_With, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
		 c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		 c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, Expires, Last-Modified, Pragma")
		 c.Header("Access-Control-Allow-Credentials", "false")
		 c.Header("Access-Control-Max-Age", "172800")
 
		 method := c.Request.Method
 		 if method == "OPTIONS" {
			 c.AbortWithStatus(http.StatusNoContent)
		 }
 
		 c.Next()
	 }
 }
 