/**
 * @Author: lidonglin
 * @Description:
 * @File:  response.go
 * @Version: 1.0.0
 * @Date: 2023/12/06 09:17
 */

package data

import (
	"encoding/json"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`

	Data interface{} `json:"data,omitempty"`
}

func MarshalData(data interface{}) string {
	retData, _ := json.Marshal(data)

	return string(retData)
}
