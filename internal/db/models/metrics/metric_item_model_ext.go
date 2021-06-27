package metrics

import "encoding/json"

func (this *MetricItem) DecodeKeys() []string {
	var result []string
	if len(this.Keys) > 0 {
		_ = json.Unmarshal([]byte(this.Keys), &result)
	}
	return result
}
