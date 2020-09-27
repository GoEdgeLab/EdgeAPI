package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type ReverseProxyService struct {
}

// 创建反向代理
func (this *ReverseProxyService) CreateReverseProxy(ctx context.Context, req *pb.CreateReverseProxyRequest) (*pb.CreateReverseProxyResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(req.SchedulingJSON, req.PrimaryOriginsJSON, req.BackupOriginsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateReverseProxyResponse{ReverseProxyId: reverseProxyId}, nil
}

// 查找反向代理
func (this *ReverseProxyService) FindEnabledReverseProxy(ctx context.Context, req *pb.FindEnabledReverseProxyRequest) (*pb.FindEnabledReverseProxyResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	reverseProxy, err := models.SharedReverseProxyDAO.FindEnabledReverseProxy(req.ReverseProxyId)
	if err != nil {
		return nil, err
	}
	if reverseProxy == nil {
		return &pb.FindEnabledReverseProxyResponse{ReverseProxy: nil}, nil
	}

	result := &pb.ReverseProxy{
		Id:                 int64(reverseProxy.Id),
		SchedulingJSON:     []byte(reverseProxy.Scheduling),
		PrimaryOriginsJSON: []byte(reverseProxy.PrimaryOrigins),
		BackupOriginsJSON:  []byte(reverseProxy.BackupOrigins),
	}
	return &pb.FindEnabledReverseProxyResponse{ReverseProxy: result}, nil
}

// 查找反向代理配置
func (this *ReverseProxyService) FindEnabledReverseProxyConfig(ctx context.Context, req *pb.FindEnabledReverseProxyConfigRequest) (*pb.FindEnabledReverseProxyConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(req.ReverseProxyId)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledReverseProxyConfigResponse{ReverseProxyJSON: configData}, nil
}

// 修改反向代理调度算法
func (this *ReverseProxyService) UpdateReverseProxyScheduling(ctx context.Context, req *pb.UpdateReverseProxySchedulingRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedReverseProxyDAO.UpdateReverseProxyScheduling(req.ReverseProxyId, req.SchedulingJSON)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改主要源站信息
func (this *ReverseProxyService) UpdateReverseProxyPrimaryOrigins(ctx context.Context, req *pb.UpdateReverseProxyPrimaryOriginsRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedReverseProxyDAO.UpdateReverseProxyPrimaryOrigins(req.ReverseProxyId, req.OriginsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改备用源站信息
func (this *ReverseProxyService) UpdateReverseProxyBackupOrigins(ctx context.Context, req *pb.UpdateReverseProxyBackupOriginsRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedReverseProxyDAO.UpdateReverseProxyBackupOrigins(req.ReverseProxyId, req.OriginsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改是否启用
func (this *ReverseProxyService) UpdateReverseProxy(ctx context.Context, req *pb.UpdateReverseProxyRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedReverseProxyDAO.UpdateReverseProxy(req.ReverseProxyId, req.RequestHost, req.RequestURI, req.StripPrefix)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}
