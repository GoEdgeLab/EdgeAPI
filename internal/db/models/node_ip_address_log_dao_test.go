package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestNodeIPAddressDAO_FindFirstNodeAccessIPAddress(t *testing.T) {
	var dao = NewNodeIPAddressDAO()
	t.Log(dao.FindFirstNodeAccessIPAddress(nil, 48, true, nodeconfigs.NodeRoleNode))
	t.Log(dao.FindFirstNodeAccessIPAddressId(nil, 48, true, nodeconfigs.NodeRoleNode))
}
