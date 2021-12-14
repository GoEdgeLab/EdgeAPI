// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewMonitorItemValueTask().Start()
		})
	})
}

// MonitorItemValueTask 节点监控数值任务
type MonitorItemValueTask struct {
}

// NewMonitorItemValueTask 获取新对象
func NewMonitorItemValueTask() *MonitorItemValueTask {
	return &MonitorItemValueTask{}
}

func (this *MonitorItemValueTask) Start() {
	ticker := time.NewTicker(24 * time.Hour)
	if Tea.IsTesting() {
		ticker = time.NewTicker(1 * time.Minute)
	}
	for range ticker.C {
		err := this.Loop()
		if err != nil {
			remotelogs.Error("MonitorItemValueTask", err.Error())
		}
	}
}

func (this *MonitorItemValueTask) Loop() error {
	return models.SharedNodeValueDAO.Clean(nil)
}
