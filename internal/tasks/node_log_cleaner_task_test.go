package tasks_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestNodeLogCleaner_loop(t *testing.T) {
	dbs.NotifyReady()

	var cleaner = tasks.NewNodeLogCleanerTask(24 * time.Hour)
	err := cleaner.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("OK")
}
