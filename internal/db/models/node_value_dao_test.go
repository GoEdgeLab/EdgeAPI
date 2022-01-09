package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/maps"
	"testing"
	"time"
)

func TestNodeValueDAO_CreateValue(t *testing.T) {
	var dao = NewNodeValueDAO()
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
	var dao = NewNodeValueDAO()
	err := dao.Clean(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
