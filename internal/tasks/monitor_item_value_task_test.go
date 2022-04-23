// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package tasks_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestMonitorItemValueTask_Loop(t *testing.T) {
	dbs.NotifyReady()

	var task = tasks.NewMonitorItemValueTask(1 * time.Minute)
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
