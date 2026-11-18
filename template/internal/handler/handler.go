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

	responseData := data.Response{
		Code:    errCode,
		Message: errMsg,
		Detail:  detail,
	}

	c.JSON(statusCode, responseData)
}

// SendPassResponse writes a standardized success response.
func SendPassResponse(c *gin.Context, respData interface{}) {
	statusCode := constant.StatusCode(constant.ErrorCodeOK)

	SendResponse(c, statusCode, respData)
}

// SendResponse writes a standard data response with the supplied HTTP status.
func SendResponse(c *gin.Context, statusCode int, respData interface{}) {
	responseData := data.Response{
		Code: constant.ErrorCodeOK,

		Data: respData,
	}

	c.JSON(statusCode, responseData)
}

// SendPassResponseEx writes a standardized success response with extended data.
func SendPassResponseEx(c *gin.Context, respData interface{}, exData interface{}) {
	statusCode := constant.StatusCode(constant.ErrorCodeOK)

	responseData := data.Response{
		Code: constant.ErrorCodeOK,

		Data:   respData,
		ExData: exData,
	}

	c.JSON(statusCode, responseData)
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
