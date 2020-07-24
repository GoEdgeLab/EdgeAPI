package tests

import (
	"context"
	"encoding/json"
	pb2 "github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	pb "github.com/TeaOSLab/EdgeAPI/internal/tests/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"
)

type server struct {
}

func (this *server) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		jsonData, _ := json.MarshalIndent(md, "", "  ")
		log.Print(string(jsonData))

		_ = md
	}

	return &pb.HelloReply{
		Message: "Hello, " + request.Name,
	}, nil
}

func TestTCPServer(t *testing.T) {
	listener, err := net.Listen("tcp", ":8001")
	if err != nil {
		t.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	pb2.RegisterNodeServiceServer(s, &services.NodeService{})

	err = s.Serve(listener)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTCPClient(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:8001", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	c := pb.NewGreeterClient(conn)

	before := time.Now()

	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "name", "liu", "age", "20")
	reply, err := c.SayHello(ctx, &pb.HelloRequest{
		Name: strings.Repeat("golang", 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(reply.Message)
	t.Log(time.Since(before).Seconds()*1000, "ms")
}

func TestTCPClient_Node(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:8001", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	c := pb2.NewNodeServiceClient(conn)

	before := time.Now()

	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "name", "liu", "age", "20")
	reply, err := c.Config(ctx, &pb2.ConfigRequest{
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(reply)
	t.Log(time.Since(before).Seconds()*1000, "ms")
}

func TestTLSServer(t *testing.T) {
	listener, err := net.Listen("tcp", ":8001")
	if err != nil {
		t.Fatal(err)
	}

	tlsCred, err := credentials.NewServerTLSFromFile("test.pem", "test.key")
	if err != nil {
		t.Fatal(err)
	}

	s := grpc.NewServer(grpc.Creds(tlsCred))
	pb.RegisterGreeterServer(s, &server{})
	err = s.Serve(listener)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTLSClient(t *testing.T) {
	tlsCred, err := credentials.NewClientTLSFromFile("test.pem", "www.hisock.cn")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := grpc.Dial("127.0.0.1:8001", grpc.WithTransportCredentials(tlsCred))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	c := pb.NewGreeterClient(conn)

	before := time.Now()

	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "name", "liu")
	reply, err := c.SayHello(ctx, &pb.HelloRequest{
		Name: strings.Repeat("golang", 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(reply.Message)
	t.Log(time.Since(before).Seconds()*1000, "ms")
}

func BenchmarkClient(b *testing.B) {
	runtime.GOMAXPROCS(1)

	tlsCred, err := credentials.NewClientTLSFromFile("test.pem", "www.hisock.cn")
	if err != nil {
		b.Fatal(err)
	}
	conn, err := grpc.Dial("127.0.0.1:8001", grpc.WithTransportCredentials(tlsCred))
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	c := pb.NewGreeterClient(conn)

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		ctx = metadata.AppendToOutgoingContext(ctx, "name", "liu")
		reply, err := c.SayHello(ctx, &pb.HelloRequest{
			Name: "golang",
		})
		_, _ = reply, err
	}
}
