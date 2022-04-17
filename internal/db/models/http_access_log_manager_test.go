// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package models_test

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestNewHTTPAccessLogManager(t *testing.T) {
	var config = &dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_log?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
		Connections: struct {
			Pool         int           `yaml:"pool"`
			Max          int           `yaml:"max"`
			Life         string        `yaml:"life"`
			LifeDuration time.Duration `yaml:",omitempty"`
		}{},
		Models: struct {
			Package string `yaml:"package"`
		}{},
	}

	db, err := dbs.NewInstanceFromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	var manager = models.SharedHTTPAccessLogManager
	err = manager.CreateTable(db, "accessLog_1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPAccessLogManager_FindTableNames(t *testing.T) {
	var config = &dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_log?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
		Connections: struct {
			Pool         int           `yaml:"pool"`
			Max          int           `yaml:"max"`
			Life         string        `yaml:"life"`
			LifeDuration time.Duration `yaml:",omitempty"`
		}{},
		Models: struct {
			Package string `yaml:"package"`
		}{},
	}

	db, err := dbs.NewInstanceFromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	for i := 0; i < 3; i++ {
		var before = time.Now()
		tables, err := models.SharedHTTPAccessLogManager.FindTables(db, "20220306")
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(tables)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}
}

func TestHTTPAccessLogManager_FindTables(t *testing.T) {
	var config = &dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_log?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
		Connections: struct {
			Pool         int           `yaml:"pool"`
			Max          int           `yaml:"max"`
			Life         string        `yaml:"life"`
			LifeDuration time.Duration `yaml:",omitempty"`
		}{},
		Models: struct {
			Package string `yaml:"package"`
		}{},
	}

	db, err := dbs.NewInstanceFromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	for i := 0; i < 3; i++ {
		var before = time.Now()
		tables, err := models.SharedHTTPAccessLogManager.FindTables(db, "20220306")
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(tables)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}
}

func TestHTTPAccessLogManager_FindLastTable(t *testing.T) {
	var config = &dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_log?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
		Connections: struct {
			Pool         int           `yaml:"pool"`
			Max          int           `yaml:"max"`
			Life         string        `yaml:"life"`
			LifeDuration time.Duration `yaml:",omitempty"`
		}{},
		Models: struct {
			Package string `yaml:"package"`
		}{},
	}

	db, err := dbs.NewInstanceFromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	for i := 0; i < 3; i++ {
		var before = time.Now()
		tableDef, err := models.SharedHTTPAccessLogManager.FindLastTable(db, "20220306", false)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(tableDef)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}
}

func TestHTTPAccessLogManager_FindPartitionTable(t *testing.T) {
	var config = &dbs.DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge_log?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
		Connections: struct {
			Pool         int           `yaml:"pool"`
			Max          int           `yaml:"max"`
			Life         string        `yaml:"life"`
			LifeDuration time.Duration `yaml:",omitempty"`
		}{},
		Models: struct {
			Package string `yaml:"package"`
		}{},
	}

	db, err := dbs.NewInstanceFromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	t.Log(models.SharedHTTPAccessLogManager.FindPartitionTable(db, timeutil.Format("Ymd", time.Now().AddDate(0, 0, -1)), -1))
	t.Log(models.SharedHTTPAccessLogManager.FindPartitionTable(db, timeutil.Format("Ymd", time.Now().AddDate(0, 0, -1)), 0))
	t.Log(models.SharedHTTPAccessLogManager.FindPartitionTable(db, timeutil.Format("Ymd", time.Now().AddDate(0, 0, -1)), 1))
}
