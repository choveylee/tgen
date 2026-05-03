// Package handler provides HTTP handlers and response helpers for the service.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/data"
)

// SendFailResponse writes a standardized failure response.
func SendFailResponse(c *gin.Context, errCode int, detail string) {
	errMsg, _ := constant.ErrMsg(errCode)
	statusCode := constant.StatusCode(errCode)

	c.JSON(statusCode, data.Response{
		Code:    errCode,
		Message: errMsg,
		Detail:  detail,
	})
}

// SendPassResponse writes a standardized success response.
func SendPassResponse(c *gin.Context, respData interface{}) {
	statusCode := constant.StatusCode(constant.ErrorCodeOK)

	c.JSON(statusCode, data.Response{
		Code: constant.ErrorCodeOK,

		Data: respData,
	})
}

// SendPassResponseEx writes a standardized success response with extended data.
func SendPassResponseEx(c *gin.Context, respData interface{}, exData interface{}) {
	statusCode := constant.StatusCode(constant.ErrorCodeOK)

	c.JSON(statusCode, data.Response{
		Code: constant.ErrorCodeOK,

		Data:   respData,
		ExData: exData,
	})
}

// SendRawJsonResponse writes respData to the response body as JSON.
func SendRawJsonResponse(c *gin.Context, respData interface{}) {
	c.JSON(http.StatusOK, respData)
}

// SendRawResponse writes respData to the response body as plain text.
func SendRawResponse(c *gin.Context, respData []byte) {
	c.Data(http.StatusOK, "text/plain", respData)
}

// RedirectUrl redirects the client to location with HTTP status 301.
func RedirectUrl(c *gin.Context, location string) {
	c.Redirect(http.StatusMovedPermanently, location)
}
