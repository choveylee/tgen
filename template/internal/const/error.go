package constant

import (
	"net/http"
)

var errorCodes = make(map[int]string)

func register(errCode int, errMsg string) int {
	if _, ok := errorCodes[errCode]; ok {
		panic("duplicate error code registration")
	}

	errorCodes[errCode] = errMsg

	return errCode
}

// ErrMsg returns the registered business error message for errCode.
func ErrMsg(errCode int) (string, bool) {
	errMsg, ok := errorCodes[errCode]

	return errMsg, ok
}

var (
	ErrorCodeOK = register(0, "")

	ErrorCodeMysqlServerAbnormal = register(100001, "Mysql服务器异常")
	ErrorCodeRedisServerAbnormal = register(100002, "Redis服务器异常")
	ErrorCodeHttpServerAbnormal  = register(100003, "Http服务器异常")

	ErrorCodeUnknownServerAbnormal = register(100011, "未知服务器异常")

	ErrorCodeRequestBodyInvalid  = register(100021, "请求Body非法")
	ErrorCodeRequestParamInvalid = register(100022, "请求参数非法")

	ErrorCodeAccessTokenInvalid = register(100031, "AccessToken非法")
	ErrorCodeAccessTokenExpired = register(100032, "AccessToken已过期")

	ErrorCodePermissionForbidden = register(100041, "权限禁止访问")
)

// StatusCode returns the HTTP status code mapped to errCode.
func StatusCode(errCode int) int {
	switch errCode {
	case ErrorCodeOK:
		return http.StatusOK

	case ErrorCodeMysqlServerAbnormal, ErrorCodeRedisServerAbnormal, ErrorCodeHttpServerAbnormal:
		return http.StatusInternalServerError

	case ErrorCodeUnknownServerAbnormal:
		return http.StatusInternalServerError

	case ErrorCodeRequestBodyInvalid, ErrorCodeRequestParamInvalid:
		return http.StatusBadRequest

	case ErrorCodeAccessTokenInvalid:
		return http.StatusUnauthorized

	case ErrorCodeAccessTokenExpired:
		return http.StatusUnauthorized

	case ErrorCodePermissionForbidden:
		return http.StatusForbidden

	default:
		panic("unrecognized error code")
	}
}
