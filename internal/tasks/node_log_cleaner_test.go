package tasks

import "testing"

func TestNodeLogCleaner_loop(t *testing.T) {
	cleaner := &NodeLogCleaner{}
	err := cleaner.loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("OK")
}
