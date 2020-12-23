package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPHeaderService struct {
	BaseService
}

// 创建Header
func (this *HTTPHeaderService) CreateHTTPHeader(ctx context.Context, req *pb.CreateHTTPHeaderRequest) (*pb.CreateHTTPHeaderResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查用户权限
	}

	headerId, err := models.SharedHTTPHeaderDAO.CreateHeader(req.Name, req.Value)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPHeaderResponse{HeaderId: headerId}, nil
}

// 修改Header
func (this *HTTPHeaderService) UpdateHTTPHeader(ctx context.Context, req *pb.UpdateHTTPHeaderRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查用户权限
	}

	err = models.SharedHTTPHeaderDAO.UpdateHeader(req.HeaderId, req.Name, req.Value)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找配置
func (this *HTTPHeaderService) FindEnabledHTTPHeaderConfig(ctx context.Context, req *pb.FindEnabledHTTPHeaderConfigRequest) (*pb.FindEnabledHTTPHeaderConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查用户权限
	}

	config, err := models.SharedHTTPHeaderDAO.ComposeHeaderConfig(req.HeaderId)
	if err != nil {
		return nil, err
	}
	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledHTTPHeaderConfigResponse{HeaderJSON: configData}, nil
}
