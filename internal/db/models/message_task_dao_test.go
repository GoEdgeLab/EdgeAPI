package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestMessageTaskDAO_CleanExpiredMessageTasks(t *testing.T) {
	var dao = models.NewMessageTaskDAO()
	var tx *dbs.Tx
	err := dao.CleanExpiredMessageTasks(tx, 30)
	if err != nil {
		t.Fatal(err)
	}
}
