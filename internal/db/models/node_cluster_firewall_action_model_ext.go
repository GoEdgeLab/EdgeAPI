package models

import (
	"encoding/json"
	"github.com/iwind/TeaGo/maps"
)

// 解析参数
func (this *NodeClusterFirewallAction) DecodeParams() (maps.Map, error) {
	if IsNotNull(this.Params) {
		params := maps.Map{}
		err := json.Unmarshal([]byte(this.Params), &params)
		if err != nil {
			return nil, err
		}
		return params, nil
	}
	return maps.Map{}, nil
}
