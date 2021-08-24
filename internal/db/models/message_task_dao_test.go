package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestMessageTaskDAO_FindSendingMessageTasks(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	tasks, err := NewMessageTaskDAO().FindSendingMessageTasks(tx, 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(tasks), "tasks")
	for _, task := range tasks {
		t.Log("task:", task.Id, "recipient:", task.RecipientId)
	}
}
