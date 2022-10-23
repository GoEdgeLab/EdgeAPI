package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestUserDAO_UpdateUserFeatures(t *testing.T) {
	var dao = models.NewUserDAO()
	var tx *dbs.Tx
	err := dao.UpdateUsersFeatures(tx, []string{
		userconfigs.UserFeatureCodeServerACME,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserDAO_CheckUserServersEnabled(t *testing.T) {
	dbs.NotifyReady()

	var dao = models.NewUserDAO()
	var tx *dbs.Tx
	for _, userId := range []int64{1, 2, 1000000} {
		isEnabled, err := dao.CheckUserServersEnabled(tx, userId)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("user:", userId, "isEnabled:", isEnabled)
	}
}
