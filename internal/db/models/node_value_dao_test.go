package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"testing"
	"time"
)

func TestNodeValueDAO_CreateValue(t *testing.T) {
	var dao = models.NewNodeValueDAO()
	m := maps.Map{
		"hello": "world12344",
	}
	err := dao.CreateValue(nil, 1, nodeconfigs.NodeRoleNode, 1, "test", m.AsJSON(), time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeValueDAO_Clean(t *testing.T) {
	var dao = models.NewNodeValueDAO()
	err := dao.Clean(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestNodeValueDAO_CreateManyValues(t *testing.T) {
	var dao = models.NewNodeValueDAO()
	var tx *dbs.Tx

	for i := 0; i < 1; i++ {
		if i%10000 == 0 {
			t.Log(i)
		}
		var item = "connections" + types.String(i)
		var clusterId int64 = 42
		var nodeId = rands.Int(1, 100)
		err := dao.CreateValue(tx, clusterId, nodeconfigs.NodeRoleNode, int64(nodeId), item, []byte(`{"total":1}`), time.Now().Unix())
		if err != nil {
			t.Fatal("item: " + item + ", err: " + err.Error())
		}
	}
	t.Log("finished")
}

func TestNodeValueDAO_SumAllNodeValues(t *testing.T) {
	var dao = models.NewNodeValueDAO()
	sum, avg, max, err := dao.SumAllNodeValues(nil, nodeconfigs.NodeRoleNode, nodeconfigs.NodeValueItemCPU, "usage", 1, nodeconfigs.NodeValueDurationUnitMinute)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("sum:", sum, "avg:", avg, "max:", max)
}

func TestNodeValueDAO_ComposeNodeStatus(t *testing.T) {
	var dao = models.NewNodeValueDAO()
	one, err := dao.Query(nil).DescPk().Find()
	if err != nil {
		t.Fatal(err)
	}

	if one != nil {
		var config = &nodeconfigs.NodeStatus{}
		err = dao.ComposeNodeStatus(nil, one.(*models.NodeValue).Role, int64(one.(*models.NodeValue).NodeId), config)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(config, t)
	}
}
