package models

import "encoding/json"

func (this *UserIdentity) DecodeFileIds() []int64 {
	if len(this.FileIds) == 0 {
		return []int64{}
	}

	var result = []int64{}
	err := json.Unmarshal(this.FileIds, &result)
	if err != nil {
		// ignore error
	}
	return result
}
