package services

import (
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestDBService_FindAllDBTables(t *testing.T) {
	db, err := dbs.Default()
	if err != nil {
		t.Fatal(err)
	}
	ones, _, err := db.FindOnes("SELECT * FROM information_schema.`TABLES` WHERE TABLE_SCHEMA=?", db.Name())
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(ones, t)
}
