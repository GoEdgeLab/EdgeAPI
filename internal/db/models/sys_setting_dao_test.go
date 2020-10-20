package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestSysSettingDAO_UpdateSetting(t *testing.T) {
	err := NewSysSettingDAO().UpdateSetting("hello", []byte(`"world"`))
	if err != nil {
		t.Fatal(err)
	}

	value, err := NewSysSettingDAO().ReadSetting("hello")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("value:", string(value))
}

func TestSysSettingDAO_UpdateSetting_Args(t *testing.T) {
	err := NewSysSettingDAO().UpdateSetting("hello %d", []byte(`"world 123"`), 123)
	if err != nil {
		t.Fatal(err)
	}

	value, err := NewSysSettingDAO().ReadSetting("hello %d", 123)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("value:", string(value))
}

func TestSysSettingDAO_CompareInt64Setting(t *testing.T) {
	i, err := NewSysSettingDAO().CompareInt64Setting("int64", 1024)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("result:", i)
}
