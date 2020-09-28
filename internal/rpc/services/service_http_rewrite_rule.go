package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

type HTTPRewriteRuleService struct {
}

// 创建重写规则
func (this *HTTPRewriteRuleService) CreateHTTPRewriteRule(ctx context.Context, req *pb.CreateHTTPRewriteRuleRequest) (*pb.CreateHTTPRewriteRuleResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	rewriteRuleId, err := models.SharedHTTPRewriteRuleDAO.CreateRewriteRule(req.Pattern, req.Replace, req.Mode, types.Int(req.RedirectStatus), req.IsBreak, req.ProxyHost, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPRewriteRuleResponse{RewriteRuleId: rewriteRuleId}, nil
}

// 修改重写规则
func (this *HTTPRewriteRuleService) UpdateHTTPRewriteRule(ctx context.Context, req *pb.UpdateHTTPRewriteRuleRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPRewriteRuleDAO.UpdateRewriteRule(req.RewriteRuleId, req.Pattern, req.Replace, req.Mode, types.Int(req.RedirectStatus), req.IsBreak, req.ProxyHost, req.IsOn)
	if err != nil {
		return nil, err
	}

	return rpcutils.RPCUpdateSuccess()
}
