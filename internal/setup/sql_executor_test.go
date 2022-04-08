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
	err := executor.Run(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestSQLExecutor_checkCluster(t *testing.T) {
	executor := NewSQLExecutor(&dbs.DBConfig{
		Driver: "mysql",
		Prefix: "edge",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_new?charset=utf8mb4&multiStatements=true",
	})
	db, err := dbs.NewInstanceFromConfig(executor.dbConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	err = executor.checkCluster(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestSQLExecutor_checkMetricItems(t *testing.T) {
	executor := NewSQLExecutor(&dbs.DBConfig{
		Driver: "mysql",
		Prefix: "edge",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_new?charset=utf8mb4&multiStatements=true",
	})
	db, err := dbs.NewInstanceFromConfig(executor.dbConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	err = executor.checkMetricItems(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestSQLExecutor_checkNS(t *testing.T) {
	executor := NewSQLExecutor(&dbs.DBConfig{
		Driver: "mysql",
		Prefix: "edge",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_new?charset=utf8mb4&multiStatements=true",
	})
	db, err := dbs.NewInstanceFromConfig(executor.dbConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	err = executor.checkNS(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
