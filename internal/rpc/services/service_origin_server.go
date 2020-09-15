package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// 源站相关管理
type OriginServerService struct {
}

// 创建源站
func (this *OriginServerService) CreateOriginServer(ctx context.Context, req *pb.CreateOriginServerRequest) (*pb.CreateOriginServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.Addr == nil {
		return nil, errors.New("'addr' can not be nil")
	}
	addrMap := maps.Map{
		"protocol":  req.Addr.Protocol,
		"portRange": req.Addr.PortRange,
		"host":      req.Addr.Host,
	}
	originId, err := models.SharedOriginServerDAO.CreateOriginServer(req.Name, string(addrMap.AsJSON()), req.Description)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOriginServerResponse{OriginId: originId}, nil
}

// 修改源站
func (this *OriginServerService) UpdateOriginServer(ctx context.Context, req *pb.UpdateOriginServerRequest) (*pb.UpdateOriginServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.Addr == nil {
		return nil, errors.New("'addr' can not be nil")
	}
	addrMap := maps.Map{
		"protocol":  req.Addr.Protocol,
		"portRange": req.Addr.PortRange,
		"host":      req.Addr.Host,
	}
	err = models.SharedOriginServerDAO.UpdateOriginServer(req.OriginId, req.Name, string(addrMap.AsJSON()), req.Description)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateOriginServerResponse{}, nil
}

// 查找单个源站信息
func (this *OriginServerService) FindEnabledOriginServer(ctx context.Context, req *pb.FindEnabledOriginServerRequest) (*pb.FindEnabledOriginServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	origin, err := models.SharedOriginServerDAO.FindEnabledOriginServer(req.OriginId)
	if err != nil {
		return nil, err
	}

	if origin == nil {
		return &pb.FindEnabledOriginServerResponse{Origin: nil}, nil
	}

	addr, err := origin.DecodeAddr()
	if err != nil {
		return nil, err
	}

	result := &pb.OriginServer{
		Id:   int64(origin.Id),
		IsOn: origin.IsOn == 1,
		Name: origin.Name,
		Addr: &pb.NetworkAddress{
			Protocol:  addr.Protocol.String(),
			Host:      addr.Host,
			PortRange: addr.PortRange,
		},
		Description: origin.Description,
	}
	return &pb.FindEnabledOriginServerResponse{Origin: result}, nil
}

// 查找源站配置
func (this *OriginServerService) FindEnabledOriginServerConfig(ctx context.Context, req *pb.FindEnabledOriginServerConfigRequest) (*pb.FindEnabledOriginServerConfigResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedOriginServerDAO.ComposeOriginConfig(req.OriginId)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledOriginServerConfigResponse{Config: configData}, nil
}
