/**
 * @Author: lidonglin
 * @Description:
 * @File:  error.go
 * @Version: 1.0.0
 * @Date: 2023/12/06 09:15
 */

package constant

import (
	"net/http"
)

var errorCodes = make(map[int]string)

func register(errCode int, errMsg string) int {
	if _, ok := errorCodes[errCode]; ok {
		panic("duplicated error code defined")
	}

	errorCodes[errCode] = errMsg

	return errCode
}

func ErrMsg(errCode int) (string, bool) {
	errMsg, ok := errorCodes[errCode]

	return errMsg, ok
}

var (
	ErrorCodeOK = register(0, "")

	ErrorCodeDbServerAbnormal   = register(100001, "database服务器异常")
	ErrorCodeRedServerAbnormal  = register(100002, "redis服务器异常")
	ErrorCodeHttpServerAbnormal = register(100003, "http服务器异常")

	ErrorCodeUnknownServerAbnormal = register(100011, "未知服务器异常")

	ErrorCodeRequestBodyIllegal  = register(100021, "请求body非法")
	ErrorCodeRequestParamIllegal = register(100022, "请求参数非法")

	ErrorCodeAccessTokenIllegal = register(100031, "access token非法")
	ErrorCodeAccessTokenExpired = register(100032, "access token已过期")

	ErrorCodePermissionForbidden = register(100041, "权限禁止")
)

func StatusCode(errCode int) int {
	switch errCode {
	case ErrorCodeOK:
		return http.StatusOK

	case ErrorCodeDbServerAbnormal, ErrorCodeRedServerAbnormal, ErrorCodeHttpServerAbnormal:
		return http.StatusInternalServerError

	case ErrorCodeUnknownServerAbnormal:
		return http.StatusInternalServerError

	case ErrorCodeRequestBodyIllegal, ErrorCodeRequestParamIllegal:
		return http.StatusBadRequest

	case ErrorCodeAccessTokenIllegal:
		return http.StatusUnauthorized

	case ErrorCodeAccessTokenExpired:
		return http.StatusUnauthorized

	case ErrorCodePermissionForbidden:
		return http.StatusForbidden

	default:
		panic("unknown error code")
	}
}
