package models

// 节点安装状态
type NodeInstallStatus struct {
	IsRunning  bool                     `json:"isRunning"`  // 是否在运行
	IsFinished bool                     `json:"isFinished"` // 是否已结束
	IsOk       bool                     `json:"isOk"`       // 是否正确安装
	Error      string                   `json:"error"`      // 错误信息
	UpdatedAt  int64                    `json:"updatedAt"`  // 更新时间，安装过程中需要每隔N秒钟更新这个状态，以便于让系统知道安装仍在进行中
	Steps      []*NodeInstallStatusStep `json:"steps"`      // 步骤
}

func NewNodeInstallStatus() *NodeInstallStatus {
	return &NodeInstallStatus{}
}
