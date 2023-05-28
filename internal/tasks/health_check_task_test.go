// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package tasks_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"testing"
	"time"
)

func TestNewHealthCheckTask(t *testing.T) {
	var task = tasks.NewHealthCheckTask(1 * time.Minute)
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
}
