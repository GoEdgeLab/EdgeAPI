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
	_, _, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	for _, nodeLog := range req.NodeLogs {
		err := models.SharedNodeLogDAO.CreateLog(tx, nodeLog.Role, nodeLog.NodeId, nodeLog.ServerId, nodeLog.OriginId, nodeLog.Level, nodeLog.Tag, nodeLog.Description, nodeLog.CreatedAt, nodeLog.Type, nodeLog.ParamsJSON)
		if err != nil {
			return nil, err
		}
	}
	return &pb.CreateNodeLogsResponse{}, nil
}

// CountNodeLogs 查询日志数量
func (this *NodeLogService) CountNodeLogs(ctx context.Context, req *pb.CountNodeLogsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedNodeLogDAO.CountNodeLogs(tx, req.Role, req.NodeClusterId, req.NodeId, req.ServerId, req.OriginId, req.AllServers, req.DayFrom, req.DayTo, req.Keyword, req.Level, types.Int8(req.FixedState), req.IsUnread, req.Tag)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListNodeLogs 列出单页日志
func (this *NodeLogService) ListNodeLogs(ctx context.Context, req *pb.ListNodeLogsRequest) (*pb.ListNodeLogsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	logs, err := models.SharedNodeLogDAO.ListNodeLogs(tx, req.Role, req.NodeClusterId, req.NodeId, req.ServerId, req.OriginId, req.AllServers, req.DayFrom, req.DayTo, req.Keyword, req.Level, types.Int8(req.FixedState), req.IsUnread, req.Tag, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var hashList = []string{}

	var result = []*pb.NodeLog{}
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
			IsFixed:     log.IsFixed,
			IsRead:      log.IsRead,
		})
	}
	return &pb.ListNodeLogsResponse{NodeLogs: result}, nil
}

// FixNodeLogs 设置日志为已修复
func (this *NodeLogService) FixNodeLogs(ctx context.Context, req *pb.FixNodeLogsRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	for _, logId := range req.NodeLogIds {
		err = models.SharedNodeLogDAO.UpdateNodeLogFixed(tx, logId)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// FixAllNodeLogs 设置所有日志为已修复
func (this *NodeLogService) FixAllNodeLogs(ctx context.Context, req *pb.FixAllNodeLogsRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeLogDAO.UpdateAllNodeLogsFixed(tx)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountAllUnreadNodeLogs 计算未读的日志数量
func (this *NodeLogService) CountAllUnreadNodeLogs(ctx context.Context, req *pb.CountAllUnreadNodeLogsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNodeLogDAO.CountAllUnreadNodeLogs(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// UpdateNodeLogsRead 设置日志为已读
func (this *NodeLogService) UpdateNodeLogsRead(ctx context.Context, req *pb.UpdateNodeLogsReadRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if len(req.NodeLogIds) > 0 {
		err = models.SharedNodeLogDAO.UpdateNodeLogIdsRead(tx, req.NodeLogIds)
		if err != nil {
			return nil, err
		}
	}

	if req.NodeId > 0 && len(req.Role) > 0 {
		err = models.SharedNodeLogDAO.UpdateNodeLogsRead(tx, req.Role, req.NodeId)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// UpdateAllNodeLogsRead 设置所有日志未已读
func (this *NodeLogService) UpdateAllNodeLogsRead(ctx context.Context, req *pb.UpdateAllNodeLogsReadRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeLogDAO.UpdateAllNodeLogsRead(tx)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
