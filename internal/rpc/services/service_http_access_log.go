package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/accesslogs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"regexp"
	"sync"
)

// HTTPAccessLogService 访问日志相关服务
type HTTPAccessLogService struct {
	BaseService
}

// CreateHTTPAccessLogs 创建访问日志
func (this *HTTPAccessLogService) CreateHTTPAccessLogs(ctx context.Context, req *pb.CreateHTTPAccessLogsRequest) (*pb.CreateHTTPAccessLogsResponse, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	if len(req.HttpAccessLogs) == 0 {
		return &pb.CreateHTTPAccessLogsResponse{}, nil
	}

	tx := this.NullTx()

	err = models.SharedHTTPAccessLogDAO.CreateHTTPAccessLogs(tx, req.HttpAccessLogs)
	if err != nil {
		return nil, err
	}

	// 发送到访问日志策略
	policyId, err := models.SharedHTTPAccessLogPolicyDAO.FindCurrentPublicPolicyId(tx)
	if err != nil {
		return nil, err
	}
	if policyId > 0 {
		err = accesslogs.SharedStorageManager.Write(policyId, req.HttpAccessLogs)
		if err != nil {
			return nil, err
		}
	}

	return &pb.CreateHTTPAccessLogsResponse{}, nil
}

// ListHTTPAccessLogs 列出单页访问日志
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

	accessLogs, requestId, hasMore, err := models.SharedHTTPAccessLogDAO.ListAccessLogs(tx, req.Partition, req.RequestId, req.Size, req.Day, req.HourFrom, req.HourTo, req.NodeClusterId, req.NodeId, req.ServerId, req.Reverse, req.HasError, req.FirewallPolicyId, req.FirewallRuleGroupId, req.FirewallRuleSetId, req.HasFirewallPolicy, req.UserId, req.Keyword, req.Ip, req.Domain)
	if err != nil {
		return nil, err
	}

	result := []*pb.HTTPAccessLog{}
	var pbNodeMap = map[int64]*pb.Node{}
	var pbClusterMap = map[int64]*pb.NodeCluster{}
	for _, accessLog := range accessLogs {
		a, err := accessLog.ToPB()
		if err != nil {
			return nil, err
		}

		// 节点 & 集群
		pbNode, ok := pbNodeMap[a.NodeId]
		if ok {
			a.Node = pbNode
		} else {
			node, err := models.SharedNodeDAO.FindEnabledNode(tx, a.NodeId)
			if err != nil {
				return nil, err
			}
			if node != nil {
				pbNode = &pb.Node{Id: int64(node.Id), Name: node.Name}

				var clusterId = int64(node.ClusterId)
				pbCluster, ok := pbClusterMap[clusterId]
				if ok {
					pbNode.NodeCluster = pbCluster
				} else {
					cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, clusterId)
					if err != nil {
						return nil, err
					}
					if cluster != nil {
						pbCluster = &pb.NodeCluster{
							Id:   int64(cluster.Id),
							Name: cluster.Name,
						}
						pbNode.NodeCluster = pbCluster
						pbClusterMap[clusterId] = pbCluster
					}
				}

				pbNodeMap[a.NodeId] = pbNode
				a.Node = pbNode
			}
		}

		result = append(result, a)
	}

	return &pb.ListHTTPAccessLogsResponse{
		HttpAccessLogs: result,
		AccessLogs:     result, // TODO 仅仅为了兼容，当用户节点版本大于0.0.8时可以删除
		HasMore:        hasMore,
		RequestId:      requestId,
	}, nil
}

// FindHTTPAccessLog 查找单个日志
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
		return &pb.FindHTTPAccessLogResponse{HttpAccessLog: nil}, nil
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
	return &pb.FindHTTPAccessLogResponse{HttpAccessLog: a}, nil
}

// FindHTTPAccessLogPartitions 查找日志分区
func (this *HTTPAccessLogService) FindHTTPAccessLogPartitions(ctx context.Context, req *pb.FindHTTPAccessLogPartitionsRequest) (*pb.FindHTTPAccessLogPartitionsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if !regexp.MustCompile(`^\d{8}$`).MatchString(req.Day) {
		return nil, errors.New("invalid 'day': " + req.Day)
	}

	var dbList = models.AllAccessLogDBs()
	if len(dbList) == 0 {
		return &pb.FindHTTPAccessLogPartitionsResponse{
			Partitions: nil,
		}, nil
	}

	var partitions = []int32{}
	var locker sync.Mutex

	var wg = sync.WaitGroup{}
	wg.Add(len(dbList))

	var lastErr error
	for _, db := range dbList {
		go func(db *dbs.DB) {
			defer wg.Done()

			names, err := models.SharedHTTPAccessLogManager.FindTableNames(db, req.Day)
			if err != nil {
				lastErr = err
			}
			for _, name := range names {
				var partition = models.SharedHTTPAccessLogManager.TablePartition(name)
				locker.Lock()
				if !lists.Contains(partitions, partition) {
					partitions = append(partitions, partition)
				}
				locker.Unlock()
			}
		}(db)
	}
	wg.Wait()

	if lastErr != nil {
		return nil, lastErr
	}

	var reversePartitions = []int32{}
	for i := len(partitions) - 1; i >= 0; i-- {
		reversePartitions = append(reversePartitions, partitions[i])
	}

	return &pb.FindHTTPAccessLogPartitionsResponse{
		Partitions:        partitions,
		ReversePartitions: reversePartitions,
	}, nil
}
