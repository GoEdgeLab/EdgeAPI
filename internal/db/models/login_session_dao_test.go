package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestLoginSessionDAO_CreateSession(t *testing.T) {
	var dao = models.NewLoginSessionDAO()
	var tx *dbs.Tx
	sessionId, err := dao.CreateSession(tx, "123456", "192.168.2.40", time.Now().Unix()+3600)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("sessionId:", sessionId)
}

func TestLoginSessionDAO_WriteSessionValue_Admin(t *testing.T) {
	var dao = models.NewLoginSessionDAO()
	var tx *dbs.Tx
	err := dao.WriteSessionValue(tx, "123456", "adminId", 123)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoginSessionDAO_WriteSessionValue_User(t *testing.T) {
	var dao = models.NewLoginSessionDAO()
	var tx *dbs.Tx
	err := dao.WriteSessionValue(tx, "123456", "userId", 123)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoginSessionDAO_WriteSessionValue(t *testing.T) {
	var dao = models.NewLoginSessionDAO()
	var tx *dbs.Tx
	err := dao.WriteSessionValue(tx, "123456", "key1", "value1")
	if err != nil {
		t.Fatal(err)
	}
}
