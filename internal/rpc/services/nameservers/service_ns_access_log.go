package nameservers

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// NSAccessLogService 访问日志相关服务
type NSAccessLogService struct {
	services.BaseService
}

// CreateNSAccessLogs 创建访问日志
func (this *NSAccessLogService) CreateNSAccessLogs(ctx context.Context, req *pb.CreateNSAccessLogsRequest) (*pb.CreateNSAccessLogsResponse, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	if len(req.NsAccessLogs) == 0 {
		return &pb.CreateNSAccessLogsResponse{}, nil
	}

	tx := this.NullTx()

	err = models.SharedNSAccessLogDAO.CreateNSAccessLogs(tx, req.NsAccessLogs)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNSAccessLogsResponse{}, nil
}

// ListNSAccessLogs 列出单页访问日志
func (this *NSAccessLogService) ListNSAccessLogs(ctx context.Context, req *pb.ListNSAccessLogsRequest) (*pb.ListNSAccessLogsResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查服务ID
	if userId > 0 {
		// TODO
	}

	accessLogs, requestId, hasMore, err := models.SharedNSAccessLogDAO.ListAccessLogs(tx, req.RequestId, req.Size, req.Day, req.NsNodeId, req.NsDomainId, req.NsRecordId, req.Keyword, req.Reverse)
	if err != nil {
		return nil, err
	}

	result := []*pb.NSAccessLog{}
	for _, accessLog := range accessLogs {
		a, err := accessLog.ToPB()
		if err != nil {
			return nil, err
		}

		// 线路
		if len(a.NsRouteCodes) > 0 {
			for _, routeCode := range a.NsRouteCodes {
				route, err := nameservers.SharedNSRouteDAO.FindEnabledRouteWithCode(nil, routeCode)
				if err != nil {
					return nil, err
				}
				if route != nil {
					a.NsRoutes = append(a.NsRoutes, &pb.NSRoute{
						Id:        types.Int64(route.Id),
						IsOn:      route.IsOn == 1,
						Name:      route.Name,
						Code:      routeCode,
						NsCluster: nil,
						NsDomain:  nil,
					})
				}
			}
		}

		result = append(result, a)
	}

	return &pb.ListNSAccessLogsResponse{
		NsAccessLogs: result,
		HasMore:      hasMore,
		RequestId:    requestId,
	}, nil
}

// FindNSAccessLog 查找单个日志
func (this *NSAccessLogService) FindNSAccessLog(ctx context.Context, req *pb.FindNSAccessLogRequest) (*pb.FindNSAccessLogResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	accessLog, err := models.SharedNSAccessLogDAO.FindAccessLogWithRequestId(tx, req.RequestId)
	if err != nil {
		return nil, err
	}
	if accessLog == nil {
		return &pb.FindNSAccessLogResponse{NsAccessLog: nil}, nil
	}

	// 检查权限
	if userId > 0 {
		// TODO
	}

	a, err := accessLog.ToPB()
	if err != nil {
		return nil, err
	}
	return &pb.FindNSAccessLogResponse{NsAccessLog: a}, nil
}
