package accesslogs

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"testing"
	"time"
)

func TestESStorage_Write(t *testing.T) {
	storage := NewESStorage(&serverconfigs.AccessLogESStorageConfig{
		Endpoint:    "http://127.0.0.1:9200",
		Index:       "logs",
		MappingType: "accessLogs",
		Username:    "hello",
		Password:    "world",
	})
	err := storage.Start()
	if err != nil {
		t.Fatal(err)
	}

	{
		err = storage.Write([]*pb.HTTPAccessLog{
			{
				RequestMethod: "POST",
				RequestPath:   "/1",
				TimeLocal:     time.Now().Format("2/Jan/2006:15:04:05 -0700"),
				TimeISO8601:   "2018-07-23T22:23:35+08:00",
				Header: map[string]*pb.Strings{
					"Content-Type": {Values: []string{"text/html"}},
				},
			},
			{
				RequestMethod: "GET",
				RequestPath:   "/2",
				TimeLocal:     time.Now().Format("2/Jan/2006:15:04:05 -0700"),
				TimeISO8601:   "2018-07-23T22:23:35+08:00",
				Header: map[string]*pb.Strings{
					"Content-Type": {Values: []string{"text/css"}},
				},
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
