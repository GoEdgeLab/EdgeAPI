package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestNodeTaskDAO_CreateNodeTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedNodeTaskDAO.CreateNodeTask(tx, 1, 2, NodeTaskTypeConfigChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeTaskDAO_CreateClusterTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedNodeTaskDAO.CreateClusterTask(tx, 1, NodeTaskTypeConfigChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeTaskDAO_ExtractClusterTask(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedNodeTaskDAO.ExtractClusterTask(tx, 1, NodeTaskTypeConfigChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
