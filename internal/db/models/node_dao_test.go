package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestNodeDAO_FindAllNodeIdsMatch(t *testing.T) {
	nodeIds, err := SharedNodeDAO.FindAllNodeIdsMatch(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(nodeIds)
}

func TestNodeDAO_FindChangedClusterIds(t *testing.T) {
	clusterIds, err := SharedNodeDAO.FindChangedClusterIds()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(clusterIds)
}

func TestNodeDAO_UpdateNodeUp(t *testing.T) {
	dbs.NotifyReady()
	isChanged, err := SharedNodeDAO.UpdateNodeUp(57, false, 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("changed:", isChanged)
}
