package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
)

// DecodeGroupIds 解析服务所属分组ID
func (this *Server) DecodeGroupIds() []int64 {
	if len(this.GroupIds) == 0 {
		return []int64{}
	}

	var result = []int64{}
	err := json.Unmarshal([]byte(this.GroupIds), &result)
	if err != nil {
		remotelogs.Error("Server.DecodeGroupIds", err.Error())
		// 忽略错误
	}
	return result
}
