package tasks_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestNodeMonitorTask_loop(t *testing.T) {
	dbs.NotifyReady()

	var task = tasks.NewNodeMonitorTask(60 * time.Second)
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeMonitorTask_Monitor(t *testing.T) {
	dbs.NotifyReady()
	var task = tasks.NewNodeMonitorTask(60 * time.Second)
	for i := 0; i < 5; i++ {
		err := task.MonitorCluster(&models.NodeCluster{
			Id: 42,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
