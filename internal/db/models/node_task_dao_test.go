package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestNodeTaskDAO_CreateNodeTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := models.SharedNodeTaskDAO.CreateNodeTask(tx, nodeconfigs.NodeRoleNode, 1, 2, 0, 0, models.NodeTaskTypeConfigChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeTaskDAO_CreateClusterTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := models.SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, 1, 0, 0, models.NodeTaskTypeConfigChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeTaskDAO_ExtractClusterTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := models.SharedNodeTaskDAO.ExtractNodeClusterTask(tx, 1, 0, 0, models.NodeTaskTypeConfigChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeTaskDAO_FindDoingNodeTasks(t *testing.T) {
	var tx *dbs.Tx
	var dao = models.NewNodeTaskDAO()
	var before = time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	_, err := dao.FindDoingNodeTasks(tx, "node", 48, 0)
	if err != nil {
		t.Fatal(err)
	}
}
