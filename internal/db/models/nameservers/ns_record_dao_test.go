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
		record := &NSRecord{RouteIds: []byte("[]")}
		t.Log(record.DecodeRouteIds())
	}

	{
		record := &NSRecord{RouteIds: []byte("[1, 2, 3]")}
		t.Log(record.DecodeRouteIds())
	}

	{
		record := &NSRecord{RouteIds: []byte(`["id:1", "id:2", "isp:liantong"]`)}
		t.Log(record.DecodeRouteIds())
	}
}
