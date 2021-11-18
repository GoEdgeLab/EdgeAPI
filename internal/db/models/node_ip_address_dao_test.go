//go:build plus
// +build plus

package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestNodeIPAddressDAO_FireThresholds(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	var nodeId int64 = 126
	node, err := SharedNodeDAO.FindEnabledNode(tx, nodeId)
	if err != nil {
		t.Fatal(err)
	}
	if node == nil {
		t.Log("node not found")
		return
	}
	err = SharedNodeIPAddressDAO.FireThresholds(tx, nodeconfigs.NodeRoleNode, nodeId)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeIPAddressDAO_LoopTasks(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedNodeIPAddressDAO.loopTask(tx, nodeconfigs.NodeRoleNode)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeIPAddressDAO_FindAddressIsHealthy(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	isHealthy, err := SharedNodeIPAddressDAO.FindAddressIsHealthy(tx, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("isHealthy:", isHealthy)
}

func TestNodeIPAddressDAO_UpdateAddressHealthCount(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	isChanged, err := SharedNodeIPAddressDAO.UpdateAddressHealthCount(tx, 1, true, 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("isChanged:", isChanged)
}
