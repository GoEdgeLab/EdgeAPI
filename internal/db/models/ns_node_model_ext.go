package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
	"time"
)

// DecodeInstallStatus 安装状态
func (this *NSNode) DecodeInstallStatus() (*NodeInstallStatus, error) {
	if len(this.InstallStatus) == 0 {
		return NewNodeInstallStatus(), nil
	}
	status := &NodeInstallStatus{}
	err := json.Unmarshal(this.InstallStatus, status)
	if err != nil {
		return NewNodeInstallStatus(), err
	}

	// 如果N秒钟没有更新状态，则认为不在运行
	if status.IsRunning && status.UpdatedAt < time.Now().Unix()-10 {
		status.IsRunning = false
		status.IsFinished = true
		status.Error = "timeout"
	}

	return status, nil
}

// DecodeStatus 节点状态
func (this *NSNode) DecodeStatus() (*nodeconfigs.NodeStatus, error) {
	if len(this.Status) == 0 {
		return nil, nil
	}
	status := &nodeconfigs.NodeStatus{}
	err := json.Unmarshal(this.Status, status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// DecodeDDoSProtection 解析DDoS Protection设置
func (this *NSNode) DecodeDDoSProtection() *ddosconfigs.ProtectionConfig {
	if IsNull(this.DdosProtection) {
		return nil
	}

	var result = &ddosconfigs.ProtectionConfig{}
	err := json.Unmarshal(this.DdosProtection, &result)
	if err != nil {
		// ignore err
	}
	return result
}

// HasDDoSProtection 检查是否有DDOS设置
func (this *NSNode) HasDDoSProtection() bool {
	var config = this.DecodeDDoSProtection()
	if config != nil {
		return !config.IsPriorEmpty()
	}
	return false
}

// DecodeConnectedAPINodes 解析连接的API节点列表
func (this *NSNode) DecodeConnectedAPINodes() []int64 {
	if IsNull(this.ConnectedAPINodes) {
		return nil
	}

	var result = []int64{}
	err := json.Unmarshal(this.ConnectedAPINodes, &result)
	if err != nil {
		// ignore err
	}
	return result
}
