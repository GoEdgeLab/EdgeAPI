package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 访问日志相关服务
type HTTPAccessLogService struct {
	BaseService
}

// 创建访问日志
func (this *HTTPAccessLogService) CreateHTTPAccessLogs(ctx context.Context, req *pb.CreateHTTPAccessLogsRequest) (*pb.CreateHTTPAccessLogsResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	if len(req.AccessLogs) == 0 {
		return &pb.CreateHTTPAccessLogsResponse{}, nil
	}

	tx := this.NullTx()

	err = models.SharedHTTPAccessLogDAO.CreateHTTPAccessLogs(tx, req.AccessLogs)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPAccessLogsResponse{}, nil
}

// 列出单页访问日志
func (this *HTTPAccessLogService) ListHTTPAccessLogs(ctx context.Context, req *pb.ListHTTPAccessLogsRequest) (*pb.ListHTTPAccessLogsResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查服务ID
	if userId > 0 {
		if req.UserId > 0 && userId != req.UserId {
			return nil, this.PermissionError()
		}

		// 这里不用担心serverId <= 0 的情况，因为如果userId>0，则只会查询当前用户下的服务，不会产生安全问题
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}
	}

	accessLogs, requestId, hasMore, err := models.SharedHTTPAccessLogDAO.ListAccessLogs(tx, req.RequestId, req.Size, req.Day, req.ServerId, req.Reverse, req.HasError, req.FirewallPolicyId, req.FirewallRuleGroupId, req.FirewallRuleSetId, req.HasFirewallPolicy, req.UserId)
	if err != nil {
		return nil, err
	}

	result := []*pb.HTTPAccessLog{}
	for _, accessLog := range accessLogs {
		a, err := accessLog.ToPB()
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}

	return &pb.ListHTTPAccessLogsResponse{
		AccessLogs: result,
		HasMore:    hasMore,
		RequestId:  requestId,
	}, nil
}

// 查找单个日志
func (this *HTTPAccessLogService) FindHTTPAccessLog(ctx context.Context, req *pb.FindHTTPAccessLogRequest) (*pb.FindHTTPAccessLogResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	accessLog, err := models.SharedHTTPAccessLogDAO.FindAccessLogWithRequestId(tx, req.RequestId)
	if err != nil {
		return nil, err
	}
	if accessLog == nil {
		return &pb.FindHTTPAccessLogResponse{AccessLog: nil}, nil
	}

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, int64(accessLog.ServerId))
		if err != nil {
			return nil, err
		}
	}

	a, err := accessLog.ToPB()
	if err != nil {
		return nil, err
	}
	return &pb.FindHTTPAccessLogResponse{AccessLog: a}, nil
}
