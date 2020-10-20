package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestMessageDAO_CreateClusterMessage(t *testing.T) {
	dao := NewMessageDAO()
	err := dao.CreateClusterMessage(1, "test", "error", "123", []byte("456"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestMessageDAO_DeleteMessagesBeforeDay(t *testing.T) {
	dao := NewMessageDAO()
	err := dao.DeleteMessagesBeforeDay(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
