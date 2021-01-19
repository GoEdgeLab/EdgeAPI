package tasks

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestServerAccessLogCleaner_Loop(t *testing.T) {
	dbs.NotifyReady()

	task := NewServerAccessLogCleaner()
	err := task.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
