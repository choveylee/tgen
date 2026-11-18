// Package data defines transport payloads used by the HTTP handlers.
package data

import (
	"encoding/json"
)

// Response is the standard envelope used by HTTP responses.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`

	Data   interface{} `json:"data,omitempty"`
	ExData interface{} `json:"ex_data,omitempty"`
}

// MarshalData marshals value to JSON and returns the encoded string.
func MarshalData(value interface{}) string {
	retData, _ := json.Marshal(value)

	return string(retData)
}
