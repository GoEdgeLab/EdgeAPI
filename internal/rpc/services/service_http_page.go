package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/regexputils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
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
	case serverconfigs.HTTPPageBodyTypeURL:
		if len(req.Url) > maxURLLength {
			return nil, errors.New("'url' too long")
		}
		if !regexputils.HTTPProtocol.MatchString(req.Url) {
			return nil, errors.New("invalid 'url' format")
		}

		if len(req.Body) > maxBodyLength { // we keep short body for user experience
			req.Body = ""
		}
	case serverconfigs.HTTPPageBodyTypeRedirectURL:
		if len(req.Url) > maxURLLength {
			return nil, errors.New("'url' too long")
		}
		if !regexputils.HTTPProtocol.MatchString(req.Url) {
			return nil, errors.New("invalid 'url' format")
		}

		if len(req.Body) > maxBodyLength { // we keep short body for user experience
			req.Body = ""
		}
	case serverconfigs.HTTPPageBodyTypeHTML:
		if len(req.Body) > maxBodyLength {
			return nil, errors.New("'body' too long")
		}

		if len(req.Url) > maxURLLength { // we keep short url for user experience
			req.Url = ""
		}
	default:
		return nil, errors.New("invalid 'bodyType': " + req.BodyType)
	}

	var exceptURLPatterns = []*shared.URLPattern{}
	if len(req.ExceptURLPatternsJSON) > 0 {
		err = json.Unmarshal(req.ExceptURLPatternsJSON, &exceptURLPatterns)
		if err != nil {
			return nil, err
		}
		for _, pattern := range exceptURLPatterns {
			err = pattern.Init()
			if err != nil {
				return nil, fmt.Errorf("validate url pattern '"+pattern.Pattern+"' failed: %w", err)
			}
		}
	}

	var onlyURLPatterns = []*shared.URLPattern{}
	if len(req.OnlyURLPatternsJSON) > 0 {
		err = json.Unmarshal(req.OnlyURLPatternsJSON, &onlyURLPatterns)
		if err != nil {
			return nil, err
		}
		for _, pattern := range onlyURLPatterns {
			err = pattern.Init()
			if err != nil {
				return nil, fmt.Errorf("validate url pattern '"+pattern.Pattern+"' failed: %w", err)
			}
		}
	}

	pageId, err := models.SharedHTTPPageDAO.CreatePage(tx, userId, req.StatusList, req.BodyType, req.Url, req.Body, types.Int(req.NewStatus), exceptURLPatterns, onlyURLPatterns)
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
	case serverconfigs.HTTPPageBodyTypeURL:
		if len(req.Url) > maxURLLength {
			return nil, errors.New("'url' too long")
		}
		if !regexputils.HTTPProtocol.MatchString(req.Url) {
			return nil, errors.New("invalid 'url' format")
		}

		if len(req.Body) > maxBodyLength { // we keep short body for user experience
			req.Body = ""
		}
	case serverconfigs.HTTPPageBodyTypeRedirectURL:
		if len(req.Url) > maxURLLength {
			return nil, errors.New("'url' too long")
		}
		if !regexputils.HTTPProtocol.MatchString(req.Url) {
			return nil, errors.New("invalid 'url' format")
		}

		if len(req.Body) > maxBodyLength { // we keep short body for user experience
			req.Body = ""
		}
	case serverconfigs.HTTPPageBodyTypeHTML:
		if len(req.Body) > maxBodyLength {
			return nil, errors.New("'body' too long")
		}

		if len(req.Url) > maxURLLength { // we keep short url for user experience
			req.Url = ""
		}
	default:
		return nil, errors.New("invalid 'bodyType': " + req.BodyType)
	}

	var exceptURLPatterns = []*shared.URLPattern{}
	if len(req.ExceptURLPatternsJSON) > 0 {
		err = json.Unmarshal(req.ExceptURLPatternsJSON, &exceptURLPatterns)
		if err != nil {
			return nil, err
		}
		for _, pattern := range exceptURLPatterns {
			err = pattern.Init()
			if err != nil {
				return nil, fmt.Errorf("validate url pattern '"+pattern.Pattern+"' failed: %w", err)
			}
		}
	}

	var onlyURLPatterns = []*shared.URLPattern{}
	if len(req.OnlyURLPatternsJSON) > 0 {
		err = json.Unmarshal(req.OnlyURLPatternsJSON, &onlyURLPatterns)
		if err != nil {
			return nil, err
		}
		for _, pattern := range onlyURLPatterns {
			err = pattern.Init()
			if err != nil {
				return nil, fmt.Errorf("validate url pattern '"+pattern.Pattern+"' failed: %w", err)
			}
		}
	}

	err = models.SharedHTTPPageDAO.UpdatePage(tx, req.HttpPageId, req.StatusList, req.BodyType, req.Url, req.Body, types.Int(req.NewStatus), exceptURLPatterns, onlyURLPatterns)
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
