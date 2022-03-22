package models

import (
	"encoding/json"
	"github.com/iwind/TeaGo/maps"
)

func (this *NodeValue) DecodeMapValue() maps.Map {
	if len(this.Value) == 0 {
		return maps.Map{}
	}
	var m = maps.Map{}
	err := json.Unmarshal(this.Value, &m)
	if err != nil {
		// 忽略错误
		return m
	}
	return m
}
