package installers

import "testing"

func TestDeployManager_LoadFiles(t *testing.T) {
	files := NewDeployManager().LoadFiles()
	for _, file := range files {
		t.Logf("%#v", file)
	}
}
