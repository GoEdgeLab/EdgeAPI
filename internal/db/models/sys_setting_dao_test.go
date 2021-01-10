package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestSysSettingDAO_UpdateSetting(t *testing.T) {
	var tx *dbs.Tx
	err := NewSysSettingDAO().UpdateSetting(tx, "hello", []byte(`"world"`))
	if err != nil {
		t.Fatal(err)
	}

	value, err := NewSysSettingDAO().ReadSetting(tx, "hello")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("value:", string(value))
}

func TestSysSettingDAO_UpdateSetting_Args(t *testing.T) {
	var tx *dbs.Tx
	err := NewSysSettingDAO().UpdateSetting(tx, "hello %d", []byte(`"world 123"`), 123)
	if err != nil {
		t.Fatal(err)
	}

	value, err := NewSysSettingDAO().ReadSetting(tx, "hello %d", 123)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("value:", string(value))
}

func TestSysSettingDAO_CompareInt64Setting(t *testing.T) {
	var tx *dbs.Tx
	i, err := NewSysSettingDAO().CompareInt64Setting(tx, "int64", 1024)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("result:", i)
}
