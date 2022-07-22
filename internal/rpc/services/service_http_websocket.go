package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPWebsocketService struct {
	BaseService
}

// 创建Websocket配置
func (this *HTTPWebsocketService) CreateHTTPWebsocket(ctx context.Context, req *pb.CreateHTTPWebsocketRequest) (*pb.CreateHTTPWebsocketResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	websocketId, err := models.SharedHTTPWebsocketDAO.CreateWebsocket(tx, req.HandshakeTimeoutJSON, req.AllowAllOrigins, req.AllowedOrigins, req.RequestSameOrigin, req.RequestOrigin)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPWebsocketResponse{WebsocketId: websocketId}, nil
}

// 修改Websocket配置
func (this *HTTPWebsocketService) UpdateHTTPWebsocket(ctx context.Context, req *pb.UpdateHTTPWebsocketRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 用户不能修改别人的WebSocket设置

	var tx = this.NullTx()

	err = models.SharedHTTPWebsocketDAO.UpdateWebsocket(tx, req.WebsocketId, req.HandshakeTimeoutJSON, req.AllowAllOrigins, req.AllowedOrigins, req.RequestSameOrigin, req.RequestOrigin)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
