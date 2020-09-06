package installers

import (
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestBaseInstaller_LookupLatest(t *testing.T) {
	installer := &BaseInstaller{}
	result, err := installer.LookupLatestInstaller("edge-node-linux-amd64")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Log("not found")
		return
	}

	t.Log("result:", result)
}
