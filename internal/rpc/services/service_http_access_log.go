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
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	if len(req.AccessLogs) == 0 {
		return &pb.CreateHTTPAccessLogsResponse{}, nil
	}

	err = models.CreateHTTPAccessLogs(req.AccessLogs)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPAccessLogsResponse{}, nil
}
