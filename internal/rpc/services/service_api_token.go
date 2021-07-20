// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// APITokenService API令牌服务
type APITokenService struct {
	BaseService
}

// FindAllEnabledAPITokens 获取API令牌
func (this *APITokenService) FindAllEnabledAPITokens(ctx context.Context, req *pb.FindAllEnabledAPITokensRequest) (*pb.FindAllEnabledAPITokensResponse, error) {
	// 这里为了安全只允许通过API节点信息获取
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAPI)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	apiTokens, err := models.SharedApiTokenDAO.FindAllEnabledAPITokens(tx, req.Role)
	if err != nil {
		return nil, err
	}
	var pbTokens = []*pb.APIToken{}
	for _, token := range apiTokens {
		pbTokens = append(pbTokens, &pb.APIToken{
			Id:     int64(token.Id),
			NodeId: token.NodeId,
			Secret: token.Secret,
			Role:   token.Role,
		})
	}
	return &pb.FindAllEnabledAPITokensResponse{ApiTokens: pbTokens}, nil
}
