package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/installers"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"time"
)

// NodeTaskService 节点同步任务相关服务
type NodeTaskService struct {
	BaseService
}

// FindNodeTasks 获取单节点同步任务
func (this *NodeTaskService) FindNodeTasks(ctx context.Context, req *pb.FindNodeTasksRequest) (*pb.FindNodeTasksResponse, error) {
	nodeType, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeNode, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	_ = req

	var tx = this.NullTx()
	tasks, err := models.SharedNodeTaskDAO.FindDoingNodeTasks(tx, nodeType, nodeId)
	if err != nil {
		return nil, err
	}

	var pbTasks = []*pb.NodeTask{}
	for _, task := range tasks {
		pbTasks = append(pbTasks, &pb.NodeTask{
			Id:        int64(task.Id),
			Type:      task.Type,
			Version:   int64(task.Version),
			IsPrimary: primaryNodeId == nodeId,
			ServerId:  int64(task.ServerId),
		})
	}

	// 边缘节点版本更新任务
	if nodeType == rpcutils.UserTypeNode && installers.SharedUpgradeLimiter.CanUpgrade() {
		status, err := models.SharedNodeDAO.FindNodeStatus(tx, nodeId)
		if err != nil {
			return nil, err
		}
		if status != nil && len(status.OS) > 0 && len(status.Arch) > 0 && len(status.BuildVersion) > 0 {
			var deployFile = installers.SharedDeployManager.FindNodeFile(status.OS, status.Arch)
			if deployFile != nil {
				if stringutil.VersionCompare(deployFile.Version, status.BuildVersion) > 0 {
					pbTasks = append(pbTasks, &pb.NodeTask{
						Type: models.NodeTaskTypeNodeVersionChanged,
					})
				}
			}
		}
	}

	return &pb.FindNodeTasksResponse{NodeTasks: pbTasks}, nil
}

// ReportNodeTaskDone 报告同步任务结果
func (this *NodeTaskService) ReportNodeTaskDone(ctx context.Context, req *pb.ReportNodeTaskDoneRequest) (*pb.RPCSuccess, error) {
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeNode, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeTaskDAO.UpdateNodeTaskDone(tx, req.NodeTaskId, req.IsOk, req.Error)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNodeClusterTasks 获取所有正在同步的集群信息
func (this *NodeTaskService) FindNodeClusterTasks(ctx context.Context, req *pb.FindNodeClusterTasksRequest) (*pb.FindNodeClusterTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	_ = req

	var tx = this.NullTx()
	// TODO 支持NS节点
	clusterIds, err := models.SharedNodeTaskDAO.FindAllDoingTaskClusterIds(tx, nodeconfigs.NodeRoleNode)
	if err != nil {
		return nil, err
	}
	if len(clusterIds) == 0 {
		return &pb.FindNodeClusterTasksResponse{ClusterTasks: []*pb.ClusterTask{}}, nil
	}

	pbClusterTasks := []*pb.ClusterTask{}
	for _, clusterId := range clusterIds {
		pbClusterTask := &pb.ClusterTask{}
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, clusterId)
		if err != nil {
			return nil, err
		}
		pbClusterTask.ClusterId = clusterId
		pbClusterTask.ClusterName = clusterName

		// 错误的节点任务
		pbNodeTasks := []*pb.NodeTask{}
		// TODO 考虑节点特别多的情形，比如只显示前100个
		// TODO 支持NS节点
		tasks, err := models.SharedNodeTaskDAO.FindAllDoingNodeTasksWithClusterId(tx, nodeconfigs.NodeRoleNode, clusterId)
		if err != nil {
			return nil, err
		}
		for _, task := range tasks {
			// 节点
			nodeName, err := models.SharedNodeDAO.FindNodeName(tx, int64(task.NodeId))
			if err != nil {
				return nil, err
			}

			// 是否超时（N秒内没有更新）
			if int64(task.UpdatedAt) < time.Now().Unix()-180 {
				task.IsDone = true
				task.IsOk = false
				task.Error = "节点响应超时"
			}

			pbNodeTasks = append(pbNodeTasks, &pb.NodeTask{
				Id:        int64(task.Id),
				Type:      task.Type,
				IsDone:    task.IsDone,
				IsOk:      task.IsOk,
				Error:     task.Error,
				UpdatedAt: int64(task.UpdatedAt),
				ServerId:  int64(task.ServerId),
				Node: &pb.Node{
					Id:   int64(task.NodeId),
					Name: nodeName,
				},
			})
		}
		pbClusterTask.NodeTasks = pbNodeTasks

		pbClusterTasks = append(pbClusterTasks, pbClusterTask)
	}

	return &pb.FindNodeClusterTasksResponse{ClusterTasks: pbClusterTasks}, nil
}

// ExistsNodeTasks 检查是否有正在执行的任务
func (this *NodeTaskService) ExistsNodeTasks(ctx context.Context, req *pb.ExistsNodeTasksRequest) (*pb.ExistsNodeTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	_ = req

	var tx = this.NullTx()

	// 是否有任务
	existTask, err := models.SharedNodeTaskDAO.ExistsDoingNodeTasks(tx, nodeconfigs.NodeRoleNode, req.ExcludeTypes)
	if err != nil {
		return nil, err
	}

	// 是否有错误
	existError, err := models.SharedNodeTaskDAO.ExistsErrorNodeTasks(tx, nodeconfigs.NodeRoleNode, req.ExcludeTypes)
	if err != nil {
		return nil, err
	}

	return &pb.ExistsNodeTasksResponse{
		ExistTasks: existTask,
		ExistError: existError,
	}, nil
}

// DeleteNodeTask 删除任务
func (this *NodeTaskService) DeleteNodeTask(ctx context.Context, req *pb.DeleteNodeTaskRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeTaskDAO.DeleteNodeTask(tx, req.NodeTaskId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteNodeTasks 批量删除任务
func (this *NodeTaskService) DeleteNodeTasks(ctx context.Context, req *pb.DeleteNodeTasksRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	for _, taskId := range req.NodeTaskIds {
		err = models.SharedNodeTaskDAO.DeleteNodeTask(tx, taskId)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// CountDoingNodeTasks 计算正在执行的任务数量
func (this *NodeTaskService) CountDoingNodeTasks(ctx context.Context, req *pb.CountDoingNodeTasksRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	_ = req

	var tx = this.NullTx()
	count, err := models.SharedNodeTaskDAO.CountDoingNodeTasks(tx, nodeconfigs.NodeRoleNode)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindNotifyingNodeTasks 查找需要通知的任务
func (this *NodeTaskService) FindNotifyingNodeTasks(ctx context.Context, req *pb.FindNotifyingNodeTasksRequest) (*pb.FindNotifyingNodeTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.Size <= 0 {
		req.Size = 100
	}
	if req.Size > 1000 {
		req.Size = 1000
	}

	var tx = this.NullTx()
	tasks, err := models.SharedNodeTaskDAO.FindNotifyingNodeTasks(tx, nodeconfigs.NodeRoleNode, req.Size)
	if err != nil {
		return nil, err
	}

	pbTasks := []*pb.NodeTask{}
	for _, task := range tasks {
		pbTasks = append(pbTasks, &pb.NodeTask{
			Id:        int64(task.Id),
			Type:      task.Type,
			IsDone:    task.IsDone,
			IsOk:      task.IsOk,
			Error:     task.Error,
			UpdatedAt: int64(task.UpdatedAt),
			Node:      &pb.Node{Id: int64(task.NodeId)},
			ServerId:  int64(task.ServerId),
		})
	}

	return &pb.FindNotifyingNodeTasksResponse{NodeTasks: pbTasks}, nil
}

// UpdateNodeTasksNotified 设置任务已通知
func (this *NodeTaskService) UpdateNodeTasksNotified(ctx context.Context, req *pb.UpdateNodeTasksNotifiedRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	err = this.RunTx(func(tx *dbs.Tx) error {
		err = models.SharedNodeTaskDAO.UpdateTasksNotified(tx, req.NodeTaskIds)
		return err
	})

	if err != nil {
		return nil, err
	}

	return this.Success()
}
