/**
 * @Author: lidonglin
 * @Description:
 * @File:  handler.go
 * @Version: 1.0.0
 * @Date: 2023/12/06 09:17
 */

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"{{domain}}/{{app_name}}/internal/const"
	"{{domain}}/{{app_name}}/internal/data"
)

func SendFailResponse(c *gin.Context, errCode int, detail string) {
	errMsg, _ := constant.ErrMsg(errCode)
	statusCode := constant.StatusCode(errCode)

	c.JSON(statusCode, data.Response{
		Code:    errCode,
		Message: errMsg,
		Detail:  detail,
	})
}

func SendPassResponse(c *gin.Context, respData interface{}) {
	statusCode := constant.StatusCode(constant.ErrorCodeOK)

	c.JSON(statusCode, data.Response{
		Code: constant.ErrorCodeOK,

		Data: respData,
	})
}

func SendPassResponseEx(c *gin.Context, respData interface{}, exData interface{}) {
    statusCode := constant.StatusCode(constant.ErrorCodeOK)

    c.JSON(statusCode, data.Response{
        Code: constant.ErrorCodeOK,

        Data:   respData,
        ExData: exData,
    })
}

func SendRawJsonResponse(c *gin.Context, respData interface{}) {
	c.JSON(http.StatusOK, respData)
}

func SendRawResponse(c *gin.Context, respData []byte) {
	c.Data(http.StatusOK, "text/plain", respData)
}

func RedirectUrl(c *gin.Context, location string) {
	c.Redirect(http.StatusMovedPermanently, location)
}
