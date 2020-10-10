package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
	"time"
)

func TestCreateHTTPAccessLogs(t *testing.T) {
	err := NewDBNodeInitializer().loop()
	if err != nil {
		t.Fatal(err)
	}

	accessLog := &pb.HTTPAccessLog{
		ServerId:  1,
		NodeId:    4,
		Status:    200,
		Timestamp: time.Now().Unix(),
	}
	dao := randomAccessLogDAO()
	t.Log("dao:", dao)
	err = CreateHTTPAccessLogsWithDAO(dao, []*pb.HTTPAccessLog{accessLog})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
