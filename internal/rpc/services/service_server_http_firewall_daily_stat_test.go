package services

import (
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestServerHTTPFirewallDailyStatService_ComposeServerHTTPFirewallDashboard(t *testing.T) {
	dbs.NotifyReady()

	service := new(ServerHTTPFirewallDailyStatService)
	resp, err := service.ComposeServerHTTPFirewallDashboard(rpcutils.NewMockAdminNodeContext(1), &pb.ComposeServerHTTPFirewallDashboardRequest{})
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(resp, t)
}
