/**
 * @Author: lidonglin
 * @Description:
 * @File:  monitor.go
 * @Version: 1.0.0
 * @Date: 2023/12/06 20:51
 */

package data

type CpuCheckData struct {
	StatusCode int `json:"-"`

	Status string `json:"Status"`
	Detail string `json:"Detail"`
}

type RamCheckData struct {
	StatusCode int `json:"-"`

	Status string `json:"Status"`
	Detail string `json:"Detail"`
}
