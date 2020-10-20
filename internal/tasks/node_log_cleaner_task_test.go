package tasks

import "testing"

func TestNodeLogCleaner_loop(t *testing.T) {
	cleaner := &NodeLogCleanerTask{}
	err := cleaner.loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("OK")
}
