package accesslogs

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/Tea"
	"testing"
	"time"
)

func TestFileStorage_Write(t *testing.T) {
	storage := NewFileStorage(&serverconfigs.AccessLogFileStorageConfig{
		Path: Tea.Root + "/logs/access-${date}.log",
	})
	err := storage.Start()
	if err != nil {
		t.Fatal(err)
	}

	{
		err = storage.Write([]*pb.HTTPAccessLog{
			{
				RequestPath: "/hello",
			},
			{
				RequestPath: "/world",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		err = storage.Write([]*pb.HTTPAccessLog{
			{
				RequestPath: "/1",
			},
			{
				RequestPath: "/2",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		err = storage.Write([]*pb.HTTPAccessLog{
			{
				RequestMethod: "POST",
				RequestPath:   "/1",
				TimeLocal:     time.Now().Format("2/Jan/2006:15:04:05 -0700"),
			},
			{
				RequestMethod: "GET",
				RequestPath:   "/2",
				TimeLocal:     time.Now().Format("2/Jan/2006:15:04:05 -0700"),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	err = storage.Close()
	if err != nil {
		t.Fatal(err)
	}
}
