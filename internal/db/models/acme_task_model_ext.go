package models

import (
	"encoding/json"
	"github.com/iwind/TeaGo/logs"
)

// 将域名解析成字符串数组
func (this *ACMETask) DecodeDomains() []string {
	if len(this.Domains) == 0 || this.Domains == "null" {
		return nil
	}
	result := []string{}
	err := json.Unmarshal([]byte(this.Domains), &result)
	if err != nil {
		logs.Error(err)
		return nil
	}
	return result
}
