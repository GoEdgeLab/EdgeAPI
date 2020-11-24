package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPHeaderService struct {
	BaseService
}

// 创建Header
func (this *HTTPHeaderService) CreateHTTPHeader(ctx context.Context, req *pb.CreateHTTPHeaderRequest) (*pb.CreateHTTPHeaderResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	headerId, err := models.SharedHTTPHeaderDAO.CreateHeader(req.Name, req.Value)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPHeaderResponse{HeaderId: headerId}, nil
}

// 修改Header
func (this *HTTPHeaderService) UpdateHTTPHeader(ctx context.Context, req *pb.UpdateHTTPHeaderRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPHeaderDAO.UpdateHeader(req.HeaderId, req.Name, req.Value)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找配置
func (this *HTTPHeaderService) FindEnabledHTTPHeaderConfig(ctx context.Context, req *pb.FindEnabledHTTPHeaderConfigRequest) (*pb.FindEnabledHTTPHeaderConfigResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
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
