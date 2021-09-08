package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/logs"
)

func (this *NodeIPAddress) DecodeThresholds() []*nodeconfigs.NodeValueThresholdConfig {
	var result = []*nodeconfigs.NodeValueThresholdConfig{}
	if len(this.Thresholds) == 0 {
		return result
	}
	err := json.Unmarshal([]byte(this.Thresholds), &result)
	if err != nil {
		// 不处理错误
		logs.Error(err)
	}
	return result
}

func (this *NodeIPAddress) DecodeConnectivity() *nodeconfigs.Connectivity {
	var connectivity = &nodeconfigs.Connectivity{}
	if len(this.Connectivity) > 0 {
		err := json.Unmarshal([]byte(this.Connectivity), connectivity)
		if err != nil {
			remotelogs.Error("NodeIPAddress.DecodeConnectivity", "decode failed: "+err.Error())
		}
	}
	return connectivity
}
