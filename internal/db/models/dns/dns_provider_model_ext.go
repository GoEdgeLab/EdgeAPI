package dns

import (
	"encoding/json"
	"github.com/iwind/TeaGo/maps"
)

// DecodeAPIParams 获取API参数
func (this *DNSProvider) DecodeAPIParams() (maps.Map, error) {
	if len(this.ApiParams) == 0 {
		return maps.Map{}, nil
	}
	result := maps.Map{}
	err := json.Unmarshal([]byte(this.ApiParams), &result)
	return result, err
}
