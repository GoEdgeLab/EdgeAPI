package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
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
	err := models.SharedIPItemDAO.NotifyUpdate(tx, 28)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestIPItemDAO_DisableIPItemsWithListId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := models.SharedIPItemDAO.DisableIPItemsWithListId(tx, 67)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestIPItemDAO_ListIPItemsAfterVersion(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	_, err := models.SharedIPItemDAO.ListIPItemsAfterVersion(tx, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestIPItemDAO_CreateManyIPs(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	var dao = models.NewIPItemDAO()
	var n = 10
	for i := 0; i < n; i++ {
		var ip = "192." + types.String(rands.Int(0, 255)) + "." + types.String(rands.Int(0, 255)) + "." + types.String(rands.Int(0, 255))
		itemId, err := dao.CreateIPItem(tx, firewallconfigs.GlobalListId, ip, ip, "", time.Now().Unix()+86400, "test", models.IPItemTypeIPv4, "warning", 0, 0, 0, 0, 0, 0, 0, false)
		if err != nil {
			t.Fatal(err)
		}
		_ = itemId
		/**err = dao.Query(tx).Pk(itemId).Attr("state", 0).UpdateQuickly()
		if err != nil {
			t.Fatal(err)
		}**/
	}
	t.Log("ok")
}

func TestIPItemDAO_DisableIPItemsWithIP(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := models.SharedIPItemDAO.DisableIPItemsWithIP(tx, "192.168.1.100", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestIPItemDAO_ParseIPValue(t *testing.T) {
	var dao = models.NewIPItemDAO()
	t.Log(dao.ParseIPValue("192.168.1.100"))
	t.Log(dao.ParseIPValue("192.168.1.100-192.168.1.200"))
	t.Log(dao.ParseIPValue("192.168.1.200-192.168.1.100"))
	t.Log(dao.ParseIPValue("192.168.1.100/24"))
	t.Log(dao.ParseIPValue("::1"))
	t.Log(dao.ParseIPValue("192.168.1.100-::2"))
	t.Log(dao.ParseIPValue("192"))
	t.Log(dao.ParseIPValue("192.168.1.200/256"))
	t.Log(dao.ParseIPValue("192.168.1.200-"))
}
