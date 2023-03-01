package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/lists"
)

// DecodeConnectivity 解析联通数值
func (this *NodeIPAddress) DecodeConnectivity() *nodeconfigs.Connectivity {
	var connectivity = &nodeconfigs.Connectivity{}
	if len(this.Connectivity) > 0 {
		err := json.Unmarshal(this.Connectivity, connectivity)
		if err != nil {
			remotelogs.Error("NodeIPAddress", "DecodeConnectivity(): decode failed: "+err.Error())
		}
	}
	return connectivity
}

// DNSIP 获取当前DNS可以使用的IP
func (this *NodeIPAddress) DNSIP() string {
	var backupIP = this.DecodeBackupIP()
	if len(backupIP) > 0 {
		return backupIP
	}
	return this.Ip
}

// DecodeBackupIP 获取备用IP
func (this *NodeIPAddress) DecodeBackupIP() string {
	if this.BackupThresholdId > 0 && len(this.BackupIP) > 0 {
		// 阈值是否存在
		b, err := SharedNodeIPAddressThresholdDAO.ExistsEnabledThreshold(nil, int64(this.BackupThresholdId))
		if err != nil {
			remotelogs.Error("NodeIPAddress", "DecodeBackupIP(): check enabled threshold failed: "+err.Error())
		} else {
			if b {
				return this.BackupIP
			}
		}
	}
	return ""
}

// DecodeClusterIds 解析集群ID
func (this *NodeIPAddress) DecodeClusterIds() []int64 {
	if IsNull(this.ClusterIds) {
		return nil
	}

	var clusterIds = []int64{}
	err := json.Unmarshal(this.ClusterIds, &clusterIds)
	if err != nil {
		remotelogs.Error("NodeIPAddress", "DecodeClusterIds(): "+err.Error())
	}
	return clusterIds
}

// IsValidInCluster 检查在某个集群中是否有效
func (this *NodeIPAddress) IsValidInCluster(clusterId int64) bool {
	var clusterIds = this.DecodeClusterIds()
	if len(clusterIds) == 0 {
		return true
	}
	return lists.ContainsInt64(clusterIds, clusterId)
}
