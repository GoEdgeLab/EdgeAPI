package models

import "encoding/json"

func (this *MetricChart) DecodeIgnoredKeys() []string {
	if len(this.IgnoredKeys) == 0 {
		return []string{}
	}

	var result = []string{}
	err := json.Unmarshal(this.IgnoredKeys, &result)
	if err != nil {
		// 这里忽略错误
		return result
	}
	return result
}
