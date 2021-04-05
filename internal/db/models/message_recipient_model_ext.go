package models

import (
	"encoding/json"
	"github.com/iwind/TeaGo/logs"
)

// 解析分组ID
func (this *MessageRecipient) DecodeGroupIds() []int64 {
	if len(this.GroupIds) == 0 {
		return []int64{}
	}
	result := []int64{}
	err := json.Unmarshal([]byte(this.GroupIds), &result)
	if err != nil {
		logs.Println("MessageRecipient.DecodeGroupIds(): " + err.Error())
		// 不阻断执行
	}
	return result
}
