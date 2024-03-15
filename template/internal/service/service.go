/**
 * @Author: lidonglin
 * @Description:
 * @File:  service.go
 * @Version: 1.0.0
 * @Date: 2023/11/15 21:49
 */

package service

import (
    "context"
 
	"github.com/choveylee/tcfg"
	"github.com/choveylee/terror"
)
 
var (
	runMode string
)
 
func InitService(ctx context.Context) *terror.Terror {
	runMode = tcfg.DefaultString(tcfg.LocalKey("RUN_MODE"), "debug")
 
	return nil
}
 