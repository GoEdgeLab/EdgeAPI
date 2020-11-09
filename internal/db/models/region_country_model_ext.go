package models

import (
	"encoding/json"
	"github.com/iwind/TeaGo/logs"
)

func (this *RegionCountry) DecodeCodes() []string {
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
