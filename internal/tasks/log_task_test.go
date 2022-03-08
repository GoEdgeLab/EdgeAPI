package tasks

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestLogTask_loopClean(t *testing.T) {
	dbs.NotifyReady()

	task := NewLogTask()
	err := task.loopClean()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestLogTask_loopMonitor(t *testing.T) {
	dbs.NotifyReady()

	task := NewLogTask()
	err := task.loopMonitor(10)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
