package models

import "encoding/json"

// DecodeKeys 解析Key
func (this *MetricStat) DecodeKeys() []string {
	var result []string
	if len(this.Keys) > 0 {
		_ = json.Unmarshal([]byte(this.Keys), &result)
	}
	return result
}
