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

	ErrorCodeMysqlServerAbnormal = register(100001, "MySQL service is unavailable")
	ErrorCodeRedisServerAbnormal = register(100002, "Redis service is unavailable")
	ErrorCodeHttpServerAbnormal  = register(100003, "HTTP service is unavailable")

	ErrorCodeUnknownServerAbnormal = register(100011, "An unexpected service error occurred")

	ErrorCodeRequestBodyIllegal  = register(100021, "The request body is invalid")
	ErrorCodeRequestParamIllegal = register(100022, "One or more request parameters are invalid")

	ErrorCodeAccessTokenIllegal = register(100031, "The access token is invalid")
	ErrorCodeAccessTokenExpired = register(100032, "The access token has expired")

	ErrorCodePermissionForbidden = register(100041, "The requested operation is not permitted")
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

	case ErrorCodeRequestBodyIllegal, ErrorCodeRequestParamIllegal:
		return http.StatusBadRequest

	case ErrorCodeAccessTokenIllegal:
		return http.StatusUnauthorized

	case ErrorCodeAccessTokenExpired:
		return http.StatusUnauthorized

	case ErrorCodePermissionForbidden:
		return http.StatusForbidden

	default:
		panic("unrecognized error code")
	}
}
