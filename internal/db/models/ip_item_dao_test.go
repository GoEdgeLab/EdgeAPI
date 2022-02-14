package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"testing"
	"time"
)

func TestIPItemDAO_NotifyClustersUpdate(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedIPItemDAO.NotifyUpdate(tx, 28)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestIPItemDAO_DisableIPItemsWithListId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedIPItemDAO.DisableIPItemsWithListId(tx, 67)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestIPItemDAO_ListIPItemsAfterVersion(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	_, err := SharedIPItemDAO.ListIPItemsAfterVersion(tx, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestIPItemDAO_CreateManyIPs(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	var dao = NewIPItemDAO()
	var n = 10
	for i := 0; i < n; i++ {
		itemId, err := dao.CreateIPItem(tx, firewallconfigs.GlobalListId, "192."+types.String(rands.Int(0, 255))+"."+types.String(rands.Int(0, 255))+"."+types.String(rands.Int(0, 255)), "", time.Now().Unix()+86400, "test", IPItemTypeIPv4, "warning", 0, 0, 0, 0, 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		_ = itemId
		/**err = dao.Query(tx).Pk(itemId).Set("state", 0).UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}**/
	}
	t.Log("ok")
}
