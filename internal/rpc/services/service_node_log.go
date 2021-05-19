package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// NodeLogService 节点日志相关服务
type NodeLogService struct {
	BaseService
}

// CreateNodeLogs 创建日志
func (this *NodeLogService) CreateNodeLogs(ctx context.Context, req *pb.CreateNodeLogsRequest) (*pb.CreateNodeLogsResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	for _, nodeLog := range req.NodeLogs {
		err := models.SharedNodeLogDAO.CreateLog(tx, nodeLog.Role, nodeLog.NodeId, nodeLog.Level, nodeLog.Tag, nodeLog.Description, nodeLog.CreatedAt)
		if err != nil {
			return nil, err
		}
	}
	return &pb.CreateNodeLogsResponse{}, nil
}

// CountNodeLogs 查询日志数量
func (this *NodeLogService) CountNodeLogs(ctx context.Context, req *pb.CountNodeLogsRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeLogDAO.CountNodeLogs(tx, req.Role, req.NodeId, req.DayFrom, req.DayTo, req.Keyword, req.Level)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListNodeLogs 列出单页日志
func (this *NodeLogService) ListNodeLogs(ctx context.Context, req *pb.ListNodeLogsRequest) (*pb.ListNodeLogsResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	logs, err := models.SharedNodeLogDAO.ListNodeLogs(tx, req.Role, req.NodeId, req.DayFrom, req.DayTo, req.Keyword, req.Level, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeLog{}
	for _, log := range logs {
		result = append(result, &pb.NodeLog{
			Role:        log.Role,
			Tag:         log.Tag,
			Description: log.Description,
			Level:       log.Level,
			NodeId:      int64(log.NodeId),
			CreatedAt:   int64(log.CreatedAt),
			Count:       types.Int32(log.Count),
		})
	}
	return &pb.ListNodeLogsResponse{NodeLogs: result}, nil
}
