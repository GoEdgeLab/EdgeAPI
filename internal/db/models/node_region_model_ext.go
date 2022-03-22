package models

import "encoding/json"

func (this *NodeRegion) DecodePriceMap() map[int64]float64 {
	var m = map[int64]float64{}
	if len(this.Prices) == 0 {
		return m
	}

	err := json.Unmarshal(this.Prices, &m)
	if err != nil {
		// 忽略错误
		return m
	}

	return m
}
