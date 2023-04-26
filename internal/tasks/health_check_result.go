package tasks

import "github.com/TeaOSLab/EdgeAPI/internal/db/models"

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Node       *models.Node
	NodeAddr   string // 节点IP地址
	NodeAddrId int64  // 节点IP地址ID
	IsOk       bool
	Error      string
	CostMs     float64
}
