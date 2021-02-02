package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

// 节点同步任务相关服务
type NodeTaskService struct {
	BaseService
}

// 获取单节点同步任务
func (this *NodeTaskService) FindNodeTasks(ctx context.Context, req *pb.FindNodeTasksRequest) (*pb.FindNodeTasksResponse, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	_ = req

	var tx = this.NullTx()
	tasks, err := models.SharedNodeTaskDAO.FindDoingNodeTasks(tx, nodeId)
	if err != nil {
		return nil, err
	}

	pbTasks := []*pb.NodeTask{}
	for _, task := range tasks {
		pbTasks = append(pbTasks, &pb.NodeTask{
			Id:   int64(task.Id),
			Type: task.Type,
		})
	}

	return &pb.FindNodeTasksResponse{NodeTasks: pbTasks}, nil
}

// 报告同步任务结果
func (this *NodeTaskService) ReportNodeTaskDone(ctx context.Context, req *pb.ReportNodeTaskDoneRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateNode(ctx)
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

// 获取所有正在同步的集群信息
func (this *NodeTaskService) FindNodeClusterTasks(ctx context.Context, req *pb.FindNodeClusterTasksRequest) (*pb.FindNodeClusterTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	_ = req

	var tx = this.NullTx()
	clusterIds, err := models.SharedNodeTaskDAO.FindAllDoingTaskClusterIds(tx)
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
		tasks, err := models.SharedNodeTaskDAO.FindAllDoingNodeTasksWithClusterId(tx, clusterId)
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
			if int64(task.UpdatedAt) < time.Now().Unix()-120 {
				task.IsDone = 1
				task.IsOk = 0
				task.Error = "节点响应超时"
			}

			pbNodeTasks = append(pbNodeTasks, &pb.NodeTask{
				Id:        int64(task.Id),
				Type:      task.Type,
				IsDone:    task.IsDone == 1,
				IsOk:      task.IsOk == 1,
				Error:     task.Error,
				UpdatedAt: int64(task.UpdatedAt),
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

// 检查是否有正在执行的任务
func (this *NodeTaskService) ExistsNodeTasks(ctx context.Context, req *pb.ExistsNodeTasksRequest) (*pb.ExistsNodeTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	_ = req

	var tx = this.NullTx()

	// 是否有任务
	existTask, err := models.SharedNodeTaskDAO.ExistsDoingNodeTasks(tx)
	if err != nil {
		return nil, err
	}

	// 是否有错误
	existError, err := models.SharedNodeTaskDAO.ExistsErrorNodeTasks(tx)
	if err != nil {
		return nil, err
	}

	return &pb.ExistsNodeTasksResponse{
		ExistTasks: existTask,
		ExistError: existError,
	}, nil
}

// 删除任务
func (this *NodeTaskService) DeleteNodeTask(ctx context.Context, req *pb.DeleteNodeTaskRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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

// 批量删除任务
func (this *NodeTaskService) DeleteNodeTasks(ctx context.Context, req *pb.DeleteNodeTasksRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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

// 计算正在执行的任务数量
func (this *NodeTaskService) CountDoingNodeTasks(ctx context.Context, req *pb.CountDoingNodeTasksRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	_ = req

	var tx = this.NullTx()
	count, err := models.SharedNodeTaskDAO.CountDoingNodeTasks(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 查找需要通知的任务
func (this *NodeTaskService) FindNotifyingNodeTasks(ctx context.Context, req *pb.FindNotifyingNodeTasksRequest) (*pb.FindNotifyingNodeTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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
	tasks, err := models.SharedNodeTaskDAO.FindNotifyingNodeTasks(tx, req.Size)
	if err != nil {
		return nil, err
	}

	pbTasks := []*pb.NodeTask{}
	for _, task := range tasks {
		pbTasks = append(pbTasks, &pb.NodeTask{
			Id:        int64(task.Id),
			Type:      task.Type,
			IsDone:    task.IsDone == 1,
			IsOk:      task.IsOk == 1,
			Error:     task.Error,
			UpdatedAt: int64(task.UpdatedAt),
			Node:      &pb.Node{Id: int64(task.NodeId)},
		})
	}

	return &pb.FindNotifyingNodeTasksResponse{NodeTasks: pbTasks}, nil
}

// 设置任务已通知
func (this *NodeTaskService) UpdateNodeTasksNotified(ctx context.Context, req *pb.UpdateNodeTasksNotifiedRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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
