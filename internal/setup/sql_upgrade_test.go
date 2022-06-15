package setup

import (
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestUpgradeSQLData(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_new?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = UpgradeSQLData(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_3_1(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_new?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_3_1(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_3_2(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_3_2(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_3_3(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_3_3(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_3_7(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_3_7(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_4_0(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_4_0(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_4_1(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_4_1(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_4_5(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_4_5(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_4_7(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_4_7(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUpgradeSQLData_v0_4_8(t *testing.T) {
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	err = upgradeV0_4_8(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
