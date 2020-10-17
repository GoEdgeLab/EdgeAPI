package tasks

import "github.com/TeaOSLab/EdgeAPI/internal/db/models"

type HealthCheckResult struct {
	Node     *models.Node
	NodeAddr string
	IsOk     bool
	Error    string
	CostMs   float64
}
