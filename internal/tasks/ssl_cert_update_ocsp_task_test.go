// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package tasks_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestSSLCertUpdateOCSPTask_Loop(t *testing.T) {
	dbs.NotifyReady()

	var task = tasks.NewSSLCertUpdateOCSPTask()
	err := task.Loop(false)
	if err != nil {
		t.Fatal(err)
	}
}
