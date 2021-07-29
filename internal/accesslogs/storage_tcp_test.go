package accesslogs

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"net"
	"testing"
	"time"
)

func TestTCPStorage_Write(t *testing.T) {
	go func() {
		server, err := net.Listen("tcp", "127.0.0.1:9981")
		if err != nil {
			t.Error(err)
			return
		}
		for {
			conn, err := server.Accept()
			if err != nil {
				break
			}

			buf := make([]byte, 1024)
			for {
				n, err := conn.Read(buf)
				if n > 0 {
					t.Log(string(buf[:n]))
				}
				if err != nil {
					break
				}
			}
			break
		}
		_ = server.Close()
	}()

	storage := NewTCPStorage(&serverconfigs.AccessLogTCPStorageConfig{
		Network: "tcp",
		Addr:    "127.0.0.1:9981",
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

	time.Sleep(2 * time.Second)

	err = storage.Close()
	if err != nil {
		t.Fatal(err)
	}
}
