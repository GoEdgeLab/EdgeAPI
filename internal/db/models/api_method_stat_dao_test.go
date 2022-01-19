package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestAPIMethodStatDAO_CreateStat(t *testing.T) {
	var dao = NewAPIMethodStatDAO()
	var tx *dbs.Tx

	err := dao.CreateStat(tx, "/pb.Hello/World", "tag", 1.123)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
