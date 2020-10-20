package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	dbs.OnReady(func() {
		go NewNodeLogCleanerTask().Start()
	})
}

// 清理节点日志的工具
type NodeLogCleanerTask struct {
	duration time.Duration
}

func NewNodeLogCleanerTask() *NodeLogCleanerTask {
	return &NodeLogCleanerTask{
		duration: 24 * time.Hour,
	}
}

func (this *NodeLogCleanerTask) Start() {
	ticker := time.NewTicker(this.duration)
	for range ticker.C {
		err := this.loop()
		if err != nil {
			logs.Println("[TASK]" + err.Error())
		}
	}
}

func (this *NodeLogCleanerTask) loop() error {
	// TODO 30天这个数值改成可以设置
	return models.SharedNodeLogDAO.DeleteExpiredLogs(30)
}
