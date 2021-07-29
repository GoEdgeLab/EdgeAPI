package accesslogs

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestStorageManager_Loop(t *testing.T) {
	dbs.NotifyReady()

	var storage = NewStorageManager()
	err := storage.Loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(storage.storageMap)
}
