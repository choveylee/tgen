/**
 * @Author: lidonglin
 * @Description:
 * @File:  monitor.go
 * @Version: 1.0.0
 * @Date: 2023/12/06 20:51
 */

package data

type CpuCheckRespData struct {
	StatusCode int `json:"-"`

	Status string `json:"Status"`
	Detail string `json:"Detail"`
}

type RamCheckRespData struct {
	StatusCode int `json:"-"`

	Status string `json:"Status"`
	Detail string `json:"Detail"`
}
