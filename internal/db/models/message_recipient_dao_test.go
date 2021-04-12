package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestMessageRecipientDAO_FindAllEnabledAndOnRecipientIdsWithGroup(t *testing.T) {
	dbs.NotifyReady()
	recipientIds, err := SharedMessageRecipientDAO.FindAllEnabledAndOnRecipientIdsWithGroup(nil, 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(recipientIds)
}
