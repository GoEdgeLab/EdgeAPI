package tasks

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHealthCheckExecutor_Run(t *testing.T) {
	dbs.NotifyReady()

	executor := NewHealthCheckExecutor(10)
	results, err := executor.Run()
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		t.Log(result.Node.Name, "addr:", result.NodeAddr, "isOk:", result.IsOk, "error:", result.Error)
	}
}
