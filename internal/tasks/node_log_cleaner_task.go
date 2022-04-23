package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewNodeLogCleanerTask(24 * time.Hour).Start()
		})
	})
}

// NodeLogCleanerTask 清理节点日志的任务
type NodeLogCleanerTask struct {
	BaseTask

	ticker *time.Ticker
}

func NewNodeLogCleanerTask(duration time.Duration) *NodeLogCleanerTask {
	return &NodeLogCleanerTask{
		ticker: time.NewTicker(duration),
	}
}

func (this *NodeLogCleanerTask) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("NodeLogCleanerTask", err.Error())
		}
	}
}

func (this *NodeLogCleanerTask) Loop() error {
	// 删除 N天 以前的info日志
	err := models.SharedNodeLogDAO.DeleteExpiredLogsWithLevel(nil, "info", 3)
	if err != nil {
		return err
	}

	// TODO 7天这个数值改成可以设置
	return models.SharedNodeLogDAO.DeleteExpiredLogs(nil, 7)
}
