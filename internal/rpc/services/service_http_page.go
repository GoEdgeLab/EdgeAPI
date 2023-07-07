package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/regexputils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
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

	// validate
	const maxURLLength = 512
	const maxBodyLength = 32 * 1024

	switch req.BodyType {
	case shared.BodyTypeURL:
		if len(req.Url) > maxURLLength {
			return nil, errors.New("'url' too long")
		}
		if !regexputils.HTTPProtocol.MatchString(req.Url) {
			return nil, errors.New("invalid 'url' format")
		}

		if len(req.Body) > maxBodyLength { // we keep short body for user experience
			req.Body = ""
		}
	case shared.BodyTypeHTML:
		if len(req.Body) > maxBodyLength {
			return nil, errors.New("'body' too long")
		}

		if len(req.Url) > maxURLLength { // we keep short url for user experience
			req.Url = ""
		}
	default:
		return nil, errors.New("invalid 'bodyType': " + req.BodyType)
	}

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

	// validate
	const maxURLLength = 512
	const maxBodyLength = 32 * 1024

	switch req.BodyType {
	case shared.BodyTypeURL:
		if len(req.Url) > maxURLLength {
			return nil, errors.New("'url' too long")
		}
		if !regexputils.HTTPProtocol.MatchString(req.Url) {
			return nil, errors.New("invalid 'url' format")
		}

		if len(req.Body) > maxBodyLength { // we keep short body for user experience
			req.Body = ""
		}
	case shared.BodyTypeHTML:
		if len(req.Body) > maxBodyLength {
			return nil, errors.New("'body' too long")
		}

		if len(req.Url) > maxURLLength { // we keep short url for user experience
			req.Url = ""
		}
	default:
		return nil, errors.New("invalid 'bodyType': " + req.BodyType)
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
