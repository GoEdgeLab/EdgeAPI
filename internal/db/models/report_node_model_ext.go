package models

import "encoding/json"

func (this *ReportNode) DecodeAllowIPs() []string {
	var result = []string{}
	if len(this.AllowIPs) > 0 {
		// 忽略错误
		_ = json.Unmarshal([]byte(this.AllowIPs), &result)
	}
	return result
}
