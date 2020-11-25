package setup

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestUpgradeSQLData(t *testing.T) {
	db, err := dbs.Default()
	if err != nil {
		t.Fatal(err)
	}
	err = UpgradeSQLData(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
