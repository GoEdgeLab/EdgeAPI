package accesslogs

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestCommandStorage_Write(t *testing.T) {
	php, err := exec.LookPath("php")
	if err != nil { // not found php, so we can not test
		t.Log("php:", err)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	before := time.Now()

	storage := NewCommandStorage(&serverconfigs.AccessLogCommandStorageConfig{
		Command: php,
		Args:    []string{cwd + "/tests/command_storage.php"},
	})
	err = storage.Start()
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Write([]*pb.HTTPAccessLog{
		{
			RequestMethod: "GET",
			RequestPath:   "/hello",
		},
		{
			RequestMethod: "GET",
			RequestPath:   "/world",
		},
		{
			RequestMethod: "GET",
			RequestPath:   "/lu",
		},
		{
			RequestMethod: "GET",
			RequestPath:   "/ping",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Close()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(time.Since(before).Seconds(), "seconds")
}
