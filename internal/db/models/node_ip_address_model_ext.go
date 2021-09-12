package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
)

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
