// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package accesslogs

import (
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"strconv"
	"time"
)

type BaseStorage struct {
	isOk         bool
	version      int
	firewallOnly bool
}

func (this *BaseStorage) SetVersion(version int) {
	this.version = version
}

func (this *BaseStorage) Version() int {
	return this.version
}

func (this *BaseStorage) IsOk() bool {
	return this.isOk
}

func (this *BaseStorage) SetOk(isOk bool) {
	this.isOk = isOk
}

func (this *BaseStorage) SetFirewallOnly(firewallOnly bool) {
	this.firewallOnly = firewallOnly
}

// Marshal 对日志进行编码
func (this *BaseStorage) Marshal(accessLog *pb.HTTPAccessLog) ([]byte, error) {
	return json.Marshal(accessLog)
}

// FormatVariables 格式化字符串中的变量
func (this *BaseStorage) FormatVariables(s string) string {
	var now = time.Now()
	return configutils.ParseVariables(s, func(varName string) (value string) {
		switch varName {
		case "year":
			return strconv.Itoa(now.Year())
		case "month":
			return fmt.Sprintf("%02d", now.Month())
		case "week":
			_, week := now.ISOWeek()
			return fmt.Sprintf("%02d", week)
		case "day":
			return fmt.Sprintf("%02d", now.Day())
		case "hour":
			return fmt.Sprintf("%02d", now.Hour())
		case "minute":
			return fmt.Sprintf("%02d", now.Minute())
		case "second":
			return fmt.Sprintf("%02d", now.Second())
		case "date":
			return fmt.Sprintf("%d%02d%02d", now.Year(), now.Month(), now.Day())
		}

		return varName
	})
}
