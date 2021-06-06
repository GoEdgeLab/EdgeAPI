package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/lists"
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
		err := models.SharedNodeLogDAO.CreateLog(tx, nodeLog.Role, nodeLog.NodeId, nodeLog.ServerId, nodeLog.Level, nodeLog.Tag, nodeLog.Description, nodeLog.CreatedAt)
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

	count, err := models.SharedNodeLogDAO.CountNodeLogs(tx, req.Role, req.NodeId, req.ServerId, req.DayFrom, req.DayTo, req.Keyword, req.Level)
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

	logs, err := models.SharedNodeLogDAO.ListNodeLogs(tx, req.Role, req.NodeId, req.ServerId, req.AllServers, req.DayFrom, req.DayTo, req.Keyword, req.Level, types.Int8(req.FixedState), req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	hashList := []string{}

	result := []*pb.NodeLog{}
	for _, log := range logs {
		// 如果是需要修复的日志，我们需要去重
		if req.FixedState > 0 {
			if lists.ContainsString(hashList, log.Hash) {
				continue
			}
			hashList = append(hashList, log.Hash)
		}

		result = append(result, &pb.NodeLog{
			Id:          int64(log.Id),
			Role:        log.Role,
			Tag:         log.Tag,
			Description: log.Description,
			Level:       log.Level,
			NodeId:      int64(log.NodeId),
			ServerId:    int64(log.ServerId),
			CreatedAt:   int64(log.CreatedAt),
			Count:       types.Int32(log.Count),
			IsFixed:     log.IsFixed == 1,
		})
	}
	return &pb.ListNodeLogsResponse{NodeLogs: result}, nil
}

// FixNodeLog 设置日志为已修复
func (this *NodeLogService) FixNodeLog(ctx context.Context, req *pb.FixNodeLogRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeLogDAO.UpdateNodeLogFixed(tx, req.NodeLogId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
