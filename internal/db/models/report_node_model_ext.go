package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
)

func (this *ReportNode) DecodeAllowIPs() []string {
	var result = []string{}
	if len(this.AllowIPs) > 0 {
		err := json.Unmarshal([]byte(this.AllowIPs), &result)
		if err != nil {
			remotelogs.Error("ReportNode.DecodeGroupIds", err.Error())
		}
	}
	return result
}

func (this *ReportNode) DecodeGroupIds() []int64 {
	var result = []int64{}
	if len(this.GroupIds) > 0 {
		err := json.Unmarshal([]byte(this.GroupIds), &result)
		if err != nil {
			remotelogs.Error("ReportNode.DecodeGroupIds", err.Error())
		}
	}
	return result
}
