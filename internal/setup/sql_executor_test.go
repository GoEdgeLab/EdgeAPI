package setup

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestSQLExecutor_Run(t *testing.T) {
	executor := NewSQLExecutor(&dbs.DBConfig{
		Driver: "mysql",
		Prefix: "edge",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_new?charset=utf8mb4&multiStatements=true",
	})
	err := executor.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
