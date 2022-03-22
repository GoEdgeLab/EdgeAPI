package models

import (
	"encoding/json"
	"github.com/iwind/TeaGo/maps"
)

// DecodeParams 解析参数
func (this *NodeClusterFirewallAction) DecodeParams() (maps.Map, error) {
	if IsNotNull(this.Params) {
		params := maps.Map{}
		err := json.Unmarshal(this.Params, &params)
		if err != nil {
			return nil, err
		}
		return params, nil
	}
	return maps.Map{}, nil
}
