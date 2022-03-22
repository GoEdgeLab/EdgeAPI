package models

import (
	"encoding/json"
)

// DecodeKeys 解析Key
func (this *MetricItem) DecodeKeys() []string {
	var result []string
	if len(this.Keys) > 0 {
		_ = json.Unmarshal(this.Keys, &result)
	}
	return result
}
