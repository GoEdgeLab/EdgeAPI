package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPWebsocketService struct {
}

// 创建Websocket配置
func (this *HTTPWebsocketService) CreateHTTPWebsocket(ctx context.Context, req *pb.CreateHTTPWebsocketRequest) (*pb.CreateHTTPWebsocketResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	websocketId, err := models.SharedHTTPWebsocketDAO.CreateWebsocket(req.HandshakeTimeoutJSON, req.AllowAllOrigins, req.AllowedOrigins, req.RequestSameOrigin, req.RequestOrigin)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPWebsocketResponse{WebsocketId: websocketId}, nil
}

// 修改Websocket配置
func (this *HTTPWebsocketService) UpdateHTTPWebsocket(ctx context.Context, req *pb.UpdateHTTPWebsocketRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPWebsocketDAO.UpdateWebsocket(req.WebsocketId, req.HandshakeTimeoutJSON, req.AllowAllOrigins, req.AllowedOrigins, req.RequestSameOrigin, req.RequestOrigin)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}
