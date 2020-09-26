package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestSysSettingDAO_UpdateSetting(t *testing.T) {
	err := SharedSysSettingDAO.UpdateSetting("hello", []byte(`"world"`))
	if err != nil {
		t.Fatal(err)
	}

	value, err := SharedSysSettingDAO.ReadSetting("hello")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("value:", string(value))
}

func TestSysSettingDAO_UpdateSetting_Args(t *testing.T) {
	err := SharedSysSettingDAO.UpdateSetting("hello %d", []byte(`"world 123"`), 123)
	if err != nil {
		t.Fatal(err)
	}

	value, err := SharedSysSettingDAO.ReadSetting("hello %d", 123)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("value:", string(value))
}
