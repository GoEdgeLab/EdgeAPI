package nameservers

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"time"
)

// DecodeInstallStatus 安装状态
func (this *NSNode) DecodeInstallStatus() (*models.NodeInstallStatus, error) {
	if len(this.InstallStatus) == 0 || this.InstallStatus == "null" {
		return models.NewNodeInstallStatus(), nil
	}
	status := &models.NodeInstallStatus{}
	err := json.Unmarshal([]byte(this.InstallStatus), status)
	if err != nil {
		return models.NewNodeInstallStatus(), err
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
	if len(this.Status) == 0 || this.Status == "null" {
		return nil, nil
	}
	status := &nodeconfigs.NodeStatus{}
	err := json.Unmarshal([]byte(this.Status), status)
	if err != nil {
		return nil, err
	}
	return status, nil
}
