//go:build plus
// +build plus

package tasks

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHealthCheckExecutor_Run(t *testing.T) {
	teaconst.IsPlus = true
	dbs.NotifyReady()

	executor := NewHealthCheckExecutor(35)
	results, err := executor.Run()
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		t.Log(result.Node.Name, "addr:", result.NodeAddr, "isOk:", result.IsOk, "error:", result.Error)
	}
}
