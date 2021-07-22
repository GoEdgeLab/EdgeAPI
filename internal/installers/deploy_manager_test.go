package installers

import "testing"

func TestDeployManager_LoadNodeFiles(t *testing.T) {
	files := NewDeployManager().LoadNodeFiles()
	for _, file := range files {
		t.Logf("%#v", file)
	}
}


func TestDeployManager_LoadNSNodeFiles(t *testing.T) {
	files := NewDeployManager().LoadNSNodeFiles()
	for _, file := range files {
		t.Logf("%#v", file)
	}
}

func TestDeployManager_FindNSNodeFile(t *testing.T) {
	file := NewDeployManager().FindNSNodeFile("linux", "amd64")
	t.Log(file)

}