package tasks_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHealthCheckClusterTask_Loop(t *testing.T) {
	dbs.NotifyReady()
	var task = tasks.NewHealthCheckClusterTask(10, nil)
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
