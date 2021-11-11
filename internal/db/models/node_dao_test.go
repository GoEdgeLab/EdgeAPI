package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestNodeDAO_FindAllNodeIdsMatch(t *testing.T) {
	var tx *dbs.Tx
	nodeIds, err := SharedNodeDAO.FindAllNodeIdsMatch(tx, 1, true, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(nodeIds)
}

func TestNodeDAO_UpdateNodeUp(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	err := SharedNodeDAO.UpdateNodeUp(tx, 57, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeDAO_FindEnabledNodeClusterIds(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	clusterIds, err := NewNodeDAO().FindEnabledAndOnNodeClusterIds(tx, 48)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(clusterIds)
}

func TestNodeDAO_ComposeNodeConfig(t *testing.T) {
	dbs.NotifyReady()

	before := time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()

	var tx *dbs.Tx
	var cacheMap = utils.NewCacheMap()
	nodeConfig, err := SharedNodeDAO.ComposeNodeConfig(tx, 48, cacheMap)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(nodeConfig.Servers), "servers")
	t.Log(cacheMap.Len(), "items")

	// old: 77ms => new: 56ms
}
