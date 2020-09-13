package services

import (
	"context"
	"encoding/base64"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/encrypt"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/assert"
	"github.com/iwind/TeaGo/maps"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

func TestAdminService_Login(t *testing.T) {
	a := assert.NewAssertion(t)
	service := &AdminService{
		debug: true,
	}
	resp, err := service.LoginAdmin(testCtx(t), &pb.LoginAdminRequest{
		Username: "admin",
		Password: stringutil.Md5("123456"),
	})
	if err != nil {
		t.Fatal(err)
	}
	a.LogJSON(resp)
}

func TestAdminService_CreateLog(t *testing.T) {
	service := &AdminService{debug: true}

	resp, err := service.CreateAdminLog(testCtx(t), &pb.CreateAdminLogRequest{
		Level:       "info",
		Description: "这是一个测试日志",
		Action:      "/login",
		Ip:          "127.0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}

func TestAdminService_FindAdminFullname(t *testing.T) {
	service := &AdminService{
		debug: true,
	}
	resp, err := service.FindAdminFullname(testCtx(t), &pb.FindAdminFullnameRequest{AdminId: 1})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}

func testCtx(t *testing.T) context.Context {
	ctx := context.Background()
	nodeId := "H6sjDf779jimnVPnBFSgZxvr6Ca0wQ0z"

	token := maps.Map{
		"timestamp": time.Now().Unix(),
		"adminId":   1,
	}
	data := token.AsJSON()

	method, err := encrypt.NewMethodInstance(teaconst.EncryptMethod, "hMHjmEng0SIcT3yiA3HIoUjogwAC9cur", nodeId)
	if err != nil {
		t.Fatal(err)
	}
	data, err = method.Encrypt(data)
	if err != nil {
		t.Fatal(err)
	}
	tokenString := base64.StdEncoding.EncodeToString(data)

	ctx = metadata.AppendToOutgoingContext(ctx, "nodeId", nodeId, "token", tokenString)
	return ctx
}
