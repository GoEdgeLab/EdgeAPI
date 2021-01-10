package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestMessageDAO_CreateClusterMessage(t *testing.T) {
	var tx *dbs.Tx

	dao := NewMessageDAO()
	err := dao.CreateClusterMessage(tx, 1, "test", "error", "123", []byte("456"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestMessageDAO_DeleteMessagesBeforeDay(t *testing.T) {
	var tx *dbs.Tx

	dao := NewMessageDAO()
	err := dao.DeleteMessagesBeforeDay(tx, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
