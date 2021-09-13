package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
)

// DecodeConnectivity 解析联通数值
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
			remotelogs.Error("NodeIPAddress.DNSIP", "check enabled threshold failed: "+err.Error())
		} else {
			if b {
				return this.BackupIP
			}
		}
	}
	return ""
}
