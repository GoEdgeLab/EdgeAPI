package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

type HTTPPageService struct {
	BaseService
}

// CreateHTTPPage 创建Page
func (this *HTTPPageService) CreateHTTPPage(ctx context.Context, req *pb.CreateHTTPPageRequest) (*pb.CreateHTTPPageResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	pageId, err := models.SharedHTTPPageDAO.CreatePage(tx, userId, req.StatusList, req.BodyType, req.Url, req.Body, types.Int(req.NewStatus))
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPPageResponse{HttpPageId: pageId}, nil
}

// UpdateHTTPPage 修改Page
func (this *HTTPPageService) UpdateHTTPPage(ctx context.Context, req *pb.UpdateHTTPPageRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedHTTPPageDAO.CheckUserPage(tx, userId, req.HttpPageId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPPageDAO.UpdatePage(tx, req.HttpPageId, req.StatusList, req.BodyType, req.Url, req.Body, types.Int(req.NewStatus))
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledHTTPPageConfig 查找单个Page配置
func (this *HTTPPageService) FindEnabledHTTPPageConfig(ctx context.Context, req *pb.FindEnabledHTTPPageConfigRequest) (*pb.FindEnabledHTTPPageConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedHTTPPageDAO.CheckUserPage(tx, userId, req.HttpPageId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPPageDAO.ComposePageConfig(tx, req.HttpPageId, nil)
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
