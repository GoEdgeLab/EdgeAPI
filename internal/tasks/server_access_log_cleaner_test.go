package tasks_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestServerAccessLogCleaner_Loop(t *testing.T) {
	dbs.NotifyReady()

	task := tasks.NewServerAccessLogCleaner()
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
