package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestNSRecord_DecodeRouteIds(t *testing.T) {
	{
		record := &NSRecord{}
		t.Log(record.DecodeRouteIds())
	}

	{
		record := &NSRecord{RouteIds: "[]"}
		t.Log(record.DecodeRouteIds())
	}

	{
		record := &NSRecord{RouteIds: "[1, 2, 3]"}
		t.Log(record.DecodeRouteIds())
	}

	{
		record := &NSRecord{RouteIds: `["id:1", "id:2", "isp:liantong"]`}
		t.Log(record.DecodeRouteIds())
	}
}
