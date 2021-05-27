package tasks

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestNodeLogCleaner_loop(t *testing.T) {
	dbs.NotifyReady()

	cleaner := &NodeLogCleanerTask{}
	err := cleaner.loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("OK")
}
