//go:build plus
// +build plus

package models_test

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"testing"
	"time"
)

func TestNodeDAO_FindAllNodeIdsMatch(t *testing.T) {
	var tx *dbs.Tx
	dbs.NotifyReady()
	nodeIds, err := models.SharedNodeDAO.FindAllNodeIdsMatch(tx, 1, true, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(nodeIds)
}

func TestNodeDAO_UpdateNodeUp(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	err := models.SharedNodeDAO.UpdateNodeUp(tx, 57, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeDAO_FindEnabledNodeClusterIds(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	clusterIds, err := models.NewNodeDAO().FindEnabledAndOnNodeClusterIds(tx, 48)
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
	nodeConfig, err := models.SharedNodeDAO.ComposeNodeConfig(tx, 48, cacheMap)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(nodeConfig.Servers), "servers")
	t.Log(cacheMap.Len(), "items")

	// old: 77ms => new: 56ms
}

func TestNodeDAO_ComposeNodeConfig_ParentNodes(t *testing.T) {
	dbs.NotifyReady()

	teaconst.IsPlus = true

	var tx *dbs.Tx
	var cacheMap = utils.NewCacheMap()
	nodeConfig, err := models.SharedNodeDAO.ComposeNodeConfig(tx, 48, cacheMap)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(nodeConfig.ParentNodes, t)
}
