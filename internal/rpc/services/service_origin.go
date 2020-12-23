package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// 源站相关管理
type OriginService struct {
	BaseService
}

// 创建源站
func (this *OriginService) CreateOrigin(ctx context.Context, req *pb.CreateOriginRequest) (*pb.CreateOriginResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
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
	originId, err := models.SharedOriginDAO.CreateOrigin(adminId, userId, req.Name, string(addrMap.AsJSON()), req.Description, req.Weight, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOriginResponse{OriginId: originId}, nil
}

// 修改源站
func (this *OriginService) UpdateOrigin(ctx context.Context, req *pb.UpdateOriginRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}
	if req.Addr == nil {
		return nil, errors.New("'addr' can not be nil")
	}
	addrMap := maps.Map{
		"protocol":  req.Addr.Protocol,
		"portRange": req.Addr.PortRange,
		"host":      req.Addr.Host,
	}
	err = models.SharedOriginDAO.UpdateOrigin(req.OriginId, req.Name, string(addrMap.AsJSON()), req.Description, req.Weight, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找单个源站信息
func (this *OriginService) FindEnabledOrigin(ctx context.Context, req *pb.FindEnabledOriginRequest) (*pb.FindEnabledOriginResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}

	origin, err := models.SharedOriginDAO.FindEnabledOrigin(req.OriginId)
	if err != nil {
		return nil, err
	}

	if origin == nil {
		return &pb.FindEnabledOriginResponse{Origin: nil}, nil
	}

	addr, err := origin.DecodeAddr()
	if err != nil {
		return nil, err
	}

	result := &pb.Origin{
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
	return &pb.FindEnabledOriginResponse{Origin: result}, nil
}

// 查找源站配置
func (this *OriginService) FindEnabledOriginConfig(ctx context.Context, req *pb.FindEnabledOriginConfigRequest) (*pb.FindEnabledOriginConfigResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}

	config, err := models.SharedOriginDAO.ComposeOriginConfig(req.OriginId)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledOriginConfigResponse{OriginJSON: configData}, nil
}
