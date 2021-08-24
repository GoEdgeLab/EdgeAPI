package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestMessageReceiverDAO_FindEnabledBestFitReceivers(t *testing.T) {
	var tx *dbs.Tx

	{
		receivers, err := NewMessageReceiverDAO().FindEnabledBestFitReceivers(tx, nodeconfigs.NodeRoleNode, 18, 1, 2, "*")
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(receivers, t)
	}

	{
		receivers, err := NewMessageReceiverDAO().FindEnabledBestFitReceivers(tx, nodeconfigs.NodeRoleNode, 30, 1, 2, "*")
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(receivers, t)
	}
}
