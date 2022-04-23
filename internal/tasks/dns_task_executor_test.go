package tasks_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestDNSTaskExecutor_Loop(t *testing.T) {
	dbs.NotifyReady()

	var task = tasks.NewDNSTaskExecutor(10 * time.Second)
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
