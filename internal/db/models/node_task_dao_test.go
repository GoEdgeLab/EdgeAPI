package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestNodeTaskDAO_CreateNodeTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedNodeTaskDAO.CreateNodeTask(tx, nodeconfigs.NodeRoleNode, 1, 2, 0, NodeTaskTypeConfigChanged, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeTaskDAO_CreateClusterTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, 1, 0, NodeTaskTypeConfigChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeTaskDAO_ExtractClusterTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedNodeTaskDAO.ExtractNodeClusterTask(tx, 1, 0, NodeTaskTypeConfigChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
