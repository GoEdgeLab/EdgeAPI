package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewNodeLogCleanerTask().Start()
		})
	})
}

// NodeLogCleanerTask 清理节点日志的任务
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
	// 删除 N天 以前的info日志
	err := models.SharedNodeLogDAO.DeleteExpiredLogsWithLevel(nil, "info", 3)
	if err != nil {
		return err
	}

	// TODO 7天这个数值改成可以设置
	return models.SharedNodeLogDAO.DeleteExpiredLogs(nil, 7)
}
