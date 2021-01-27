package tasks

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestDNSTaskExecutor_Loop(t *testing.T) {
	dbs.NotifyReady()

	executor := NewDNSTaskExecutor()
	err := executor.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
