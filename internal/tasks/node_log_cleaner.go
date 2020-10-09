package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	go NewNodeLogCleaner().Start()
}

// 清理节点日志的工具
type NodeLogCleaner struct {
	duration time.Duration
}

func NewNodeLogCleaner() *NodeLogCleaner {
	return &NodeLogCleaner{
		duration: 24 * time.Hour,
	}
}

func (this *NodeLogCleaner) Start() {
	ticker := time.NewTicker(this.duration)
	for range ticker.C {
		err := this.loop()
		if err != nil {
			logs.Println("[TASK]" + err.Error())
		}
	}
}

func (this *NodeLogCleaner) loop() error {
	// TODO 30天这个数值改成可以设置
	return models.SharedNodeLogDAO.DeleteExpiredLogs(30)
}
