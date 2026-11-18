// Package handler provides HTTP handlers and response helpers for the service.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/data"
)

// SendFailResponse writes a standardized failure response.
func SendFailResponse(ginContext *gin.Context, errCode int, detail string) {
	errMsg, ok := constant.ErrMsg(errCode)
	if !ok || errCode == constant.ErrorCodeOK {
		errCode = constant.ErrorCodeUnknownServerAbnormal
		errMsg, _ = constant.ErrMsg(errCode)
	}

	statusCode := constant.StatusCode(errCode)

	responseData := data.Response{
		Code:    errCode,
		Message: errMsg,
		Detail:  detail,
	}

	ginContext.JSON(statusCode, responseData)
}

// SendPassResponse writes a standardized success response.
func SendPassResponse(ginContext *gin.Context, respData interface{}) {
	statusCode := constant.StatusCode(constant.ErrorCodeOK)

	SendResponse(ginContext, statusCode, respData)
}

// SendResponse writes a standard data response with the supplied HTTP status.
func SendResponse(ginContext *gin.Context, statusCode int, respData interface{}) {
	responseData := data.Response{
		Code: constant.ErrorCodeOK,

		Data: respData,
	}

	ginContext.JSON(statusCode, responseData)
}

// SendPassResponseEx writes a standardized success response with extended data.
func SendPassResponseEx(ginContext *gin.Context, respData interface{}, exData interface{}) {
	statusCode := constant.StatusCode(constant.ErrorCodeOK)

	responseData := data.Response{
		Code: constant.ErrorCodeOK,

		Data:   respData,
		ExData: exData,
	}

	ginContext.JSON(statusCode, responseData)
}

// SendRawJsonResponse writes respData to the response body as JSON.
func SendRawJsonResponse(ginContext *gin.Context, respData interface{}) {
	ginContext.JSON(http.StatusOK, respData)
}

// SendRawResponse writes respData to the response body as plain text.
func SendRawResponse(ginContext *gin.Context, respData []byte) {
	ginContext.Data(http.StatusOK, "text/plain", respData)
}

// RedirectUrl redirects the client to location with HTTP status 301.
func RedirectUrl(ginContext *gin.Context, location string) {
	ginContext.Redirect(http.StatusMovedPermanently, location)
}
