package models

import (
	"encoding/json"
	"time"
)

// 安装状态
func (this *Node) DecodeInstallStatus() (*NodeInstallStatus, error) {
	if len(this.InstallStatus) == 0 || this.InstallStatus == "null" {
		return NewNodeInstallStatus(), nil
	}
	status := &NodeInstallStatus{}
	err := json.Unmarshal([]byte(this.InstallStatus), status)
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

