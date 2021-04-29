// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package tasks

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestMonitorItemValueTask_Loop(t *testing.T) {
	dbs.NotifyReady()

	task := NewMonitorItemValueTask()
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
