package models

import (
	_ "github.com/go-sql-driver/mysql"
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
