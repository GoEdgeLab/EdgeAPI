package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 访问日志相关服务
type HTTPAccessLogService struct {
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

	err = models.SharedHTTPAccessLogDAO.CreateHTTPAccessLogs(req.AccessLogs)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPAccessLogsResponse{}, nil
}

// 列出单页访问日志
func (this *HTTPAccessLogService) ListHTTPAccessLogs(ctx context.Context, req *pb.ListHTTPAccessLogsRequest) (*pb.ListHTTPAccessLogsResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	accessLogs, requestId, hasMore, err := models.SharedHTTPAccessLogDAO.ListAccessLogs(req.RequestId, req.Size, req.Day, req.ServerId, req.Reverse)
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
