package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// HTTPRewriteRuleService 重写规则相关服务
type HTTPRewriteRuleService struct {
	BaseService
}

// CreateHTTPRewriteRule 创建重写规则
func (this *HTTPRewriteRuleService) CreateHTTPRewriteRule(ctx context.Context, req *pb.CreateHTTPRewriteRuleRequest) (*pb.CreateHTTPRewriteRuleResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	rewriteRuleId, err := models.SharedHTTPRewriteRuleDAO.CreateRewriteRule(tx, req.Pattern, req.Replace, req.Mode, types.Int(req.RedirectStatus), req.IsBreak, req.ProxyHost, req.WithQuery, req.IsOn, req.CondsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPRewriteRuleResponse{RewriteRuleId: rewriteRuleId}, nil
}

// UpdateHTTPRewriteRule 修改重写规则
func (this *HTTPRewriteRuleService) UpdateHTTPRewriteRule(ctx context.Context, req *pb.UpdateHTTPRewriteRuleRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedHTTPRewriteRuleDAO.UpdateRewriteRule(tx, req.RewriteRuleId, req.Pattern, req.Replace, req.Mode, types.Int(req.RedirectStatus), req.IsBreak, req.ProxyHost, req.WithQuery, req.IsOn, req.CondsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
