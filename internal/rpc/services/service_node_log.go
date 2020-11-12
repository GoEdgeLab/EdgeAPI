package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 节点日志相关服务
type NodeLogService struct {
}

// 创建日志
func (this *NodeLogService) CreateNodeLogs(ctx context.Context, req *pb.CreateNodeLogsRequest) (*pb.CreateNodeLogsResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	for _, nodeLog := range req.NodeLogs {
		err := models.SharedNodeLogDAO.CreateLog(nodeLog.Role, nodeLog.NodeId, nodeLog.Level, nodeLog.Tag, nodeLog.Description, nodeLog.CreatedAt)
		if err != nil {
			return nil, err
		}
	}
	return &pb.CreateNodeLogsResponse{}, nil
}

// 查询日志数量
func (this *NodeLogService) CountNodeLogs(ctx context.Context, req *pb.CountNodeLogsRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedNodeLogDAO.CountNodeLogs(req.Role, req.NodeId)
	if err != nil {
		return nil, err
	}
	return &pb.RPCCountResponse{Count: count}, nil
}

// 列出单页日志
func (this *NodeLogService) ListNodeLogs(ctx context.Context, req *pb.ListNodeLogsRequest) (*pb.ListNodeLogsResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	logs, err := models.SharedNodeLogDAO.ListNodeLogs(req.Role, req.NodeId, req.Offset, req.Size)
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
		})
	}
	return &pb.ListNodeLogsResponse{NodeLogs: result}, nil
}
