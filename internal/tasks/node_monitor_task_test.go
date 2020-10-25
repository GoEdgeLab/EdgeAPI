package tasks

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestNodeMonitorTask_loop(t *testing.T) {
	dbs.NotifyReady()

	task := NewNodeMonitorTask(60)
	err := task.loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
