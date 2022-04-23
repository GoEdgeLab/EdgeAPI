// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewMonitorItemValueTask(1 * time.Hour).Start()
		})
	})
}

// MonitorItemValueTask 节点监控数值任务
type MonitorItemValueTask struct {
	BaseTask

	ticker *time.Ticker
}

// NewMonitorItemValueTask 获取新对象
func NewMonitorItemValueTask(duration time.Duration) *MonitorItemValueTask {
	var ticker = time.NewTicker(duration)
	if Tea.IsTesting() {
		ticker = time.NewTicker(1 * time.Minute)
	}

	return &MonitorItemValueTask{
		ticker: ticker,
	}
}

func (this *MonitorItemValueTask) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("MonitorItemValueTask", err.Error())
		}
	}
}

func (this *MonitorItemValueTask) Loop() error {
	return models.SharedNodeValueDAO.Clean(nil)
}
