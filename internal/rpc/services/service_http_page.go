package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

type HTTPPageService struct {
}

// 创建Page
func (this *HTTPPageService) CreateHTTPPage(ctx context.Context, req *pb.CreateHTTPPageRequest) (*pb.CreateHTTPPageResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	pageId, err := models.SharedHTTPPageDAO.CreatePage(req.StatusList, req.Url, types.Int(req.NewStatus))
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPPageResponse{PageId: pageId}, nil
}

// 修改Page
func (this *HTTPPageService) UpdateHTTPPage(ctx context.Context, req *pb.UpdateHTTPPageRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPPageDAO.UpdatePage(req.PageId, req.StatusList, req.Url, types.Int(req.NewStatus))
	if err != nil {
		return nil, err
	}

	return rpcutils.Success()
}

// 查找单个Page配置
func (this *HTTPPageService) FindEnabledHTTPPageConfig(ctx context.Context, req *pb.FindEnabledHTTPPageConfigRequest) (*pb.FindEnabledHTTPPageConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedHTTPPageDAO.ComposePageConfig(req.PageId)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledHTTPPageConfigResponse{
		PageJSON: configJSON,
	}, nil
}
