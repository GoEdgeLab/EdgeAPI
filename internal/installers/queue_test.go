package installers

import (
	"testing"
	"time"
)

func TestQueue_InstallNode(t *testing.T) {
	queue := NewQueue()
	err := queue.InstallNodeProcess(16, false)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	t.Log("OK")

}
