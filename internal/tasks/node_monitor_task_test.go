package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestNodeMonitorTask_loop(t *testing.T) {
	dbs.NotifyReady()

	var task = NewNodeMonitorTask(60)
	err := task.loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeMonitorTask_Monitor(t *testing.T) {
	dbs.NotifyReady()
	var task = NewNodeMonitorTask(60)
	for i := 0; i < 5; i++ {
		err := task.monitorCluster(&models.NodeCluster{
			Id: 42,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
