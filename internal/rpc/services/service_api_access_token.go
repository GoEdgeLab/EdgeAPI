package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// APIAccessTokenService AccessToken相关服务
type APIAccessTokenService struct {
	BaseService
}

// GetAPIAccessToken 获取AccessToken
func (this *APIAccessTokenService) GetAPIAccessToken(ctx context.Context, req *pb.GetAPIAccessTokenRequest) (*pb.GetAPIAccessTokenResponse, error) {
	if req.Type == "user" { // 用户
		tx := this.NullTx()

		accessKey, err := models.SharedUserAccessKeyDAO.FindAccessKeyWithUniqueId(tx, req.AccessKeyId)
		if err != nil {
			return nil, err
		}
		if accessKey == nil {
			return nil, errors.New("access key not found")
		}
		if accessKey.Secret != req.AccessKey {
			return nil, errors.New("access key not found")
		}

		// 更新AccessKey访问时间
		err = models.SharedUserAccessKeyDAO.UpdateAccessKeyAccessedAt(tx, int64(accessKey.Id))
		if err != nil {
			return nil, err
		}

		// 创建AccessToken
		token, expiresAt, err := models.SharedAPIAccessTokenDAO.GenerateAccessToken(tx, int64(accessKey.UserId))
		if err != nil {
			return nil, err
		}

		return &pb.GetAPIAccessTokenResponse{
			Token:     token,
			ExpiresAt: expiresAt,
		}, nil
	} else {
		return nil, errors.New("unsupported type '" + req.Type + "'")
	}
}
