package services

import (
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestServerService_UploadServerHTTPRequestStat(t *testing.T) {
	dbs.NotifyReady()

	service := new(ServerService)
	_, err := service.UploadServerHTTPRequestStat(rpcutils.NewMockNodeContext(1), &pb.UploadServerHTTPRequestStatRequest{
		Month: timeutil.Format("Ym"),
		RegionCities: []*pb.UploadServerHTTPRequestStatRequest_RegionCity{
			{
				ServerId:     1,
				CountryName:  "中国",
				ProvinceName: "安徽省",
				CityName:     "阜阳市",
				Count:        1,
			},
		},
		RegionProviders: []*pb.UploadServerHTTPRequestStatRequest_RegionProvider{
			{
				ServerId: 1,
				Name:     "电信",
				Count:    1,
			},
		},
		Systems: []*pb.UploadServerHTTPRequestStatRequest_System{
			{
				ServerId: 1,
				Name:     "Mac OS X",
				Count:    1,
				Version:  "20",
			},
		},
		Browsers: []*pb.UploadServerHTTPRequestStatRequest_Browser{
			{
				ServerId: 1,
				Name:     "Chrome",
				Count:    1,
				Version:  "70",
			},
			{
				ServerId: 1,
				Name:     "Firefox",
				Count:    1,
				Version:  "30",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("===countries===")
	logs.PrintAsJSON(serverHTTPCountryStatMap, t)

	t.Log("===provinces===")
	logs.PrintAsJSON(serverHTTPProvinceStatMap, t)

	t.Log("===cities===")
	logs.PrintAsJSON(serverHTTPCityStatMap, t)

	t.Log("===providers===")
	logs.PrintAsJSON(serverHTTPProviderStatMap, t)

	t.Log("===systems===")
	logs.PrintAsJSON(serverHTTPSystemStatMap, t)

	t.Log("===browsers===")
	logs.PrintAsJSON(serverHTTPBrowserStatMap, t)

	err = service.dumpServerHTTPStats()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
