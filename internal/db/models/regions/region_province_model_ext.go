package regions

import (
	"encoding/json"
	"github.com/iwind/TeaGo/logs"
)

func (this *RegionProvince) DecodeCodes() []string {
	if len(this.Codes) == 0 {
		return []string{}
	}
	result := []string{}
	err := json.Unmarshal([]byte(this.Codes), &result)
	if err != nil {
		logs.Error(err)
	}
	return result
}
