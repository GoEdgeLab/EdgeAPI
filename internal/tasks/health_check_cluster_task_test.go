package tasks

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHealthCheckClusterTask_loop(t *testing.T) {
	dbs.NotifyReady()
	task := NewHealthCheckClusterTask(10, nil)
	err := task.loop(10)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
