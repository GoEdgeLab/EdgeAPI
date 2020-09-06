package utils

import (
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestUnzip_Run(t *testing.T) {
	unzip := NewUnzip(Tea.Root+"/deploy/edge-node-v0.0.1.zip", Tea.Root+"/deploy/")
	err := unzip.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("OK")
}
