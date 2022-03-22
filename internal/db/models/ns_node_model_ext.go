package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
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
