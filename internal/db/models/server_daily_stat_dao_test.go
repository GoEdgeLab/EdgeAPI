package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestServerDailyStatDAO_SaveStats(t *testing.T) {
	stats := []*pb.ServerDailyStat{
		{
			ServerId:  1,
			RegionId:  2,
			Bytes:     1,
			CreatedAt: 1607671488,
		},
	}
	err := NewServerDailyStatDAO().SaveStats(stats)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestServerDailyStatDAO_SaveStats2(t *testing.T) {
	stats := []*pb.ServerDailyStat{
		{
			ServerId:  1,
			RegionId:  3,
			Bytes:     1,
			CreatedAt: 1607671488,
		},
	}
	err := NewServerDailyStatDAO().SaveStats(stats)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
