package services

import (
	"context"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// DNSTaskService DNS同步相关任务
type DNSTaskService struct {
	BaseService
}

// ExistsDNSTasks 检查是否有正在执行的任务
func (this *DNSTaskService) ExistsDNSTasks(ctx context.Context, req *pb.ExistsDNSTasksRequest) (*pb.ExistsDNSTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	existDoingTasks, err := dns.SharedDNSTaskDAO.ExistDoingTasks(tx)
	if err != nil {
		return nil, err
	}

	existErrorTasks, err := dns.SharedDNSTaskDAO.ExistErrorTasks(tx)
	if err != nil {
		return nil, err
	}

	return &pb.ExistsDNSTasksResponse{
		ExistTasks: existDoingTasks,
		ExistError: existErrorTasks,
	}, nil
}

// FindAllDoingDNSTasks 查找正在执行的所有任务
func (this *DNSTaskService) FindAllDoingDNSTasks(ctx context.Context, req *pb.FindAllDoingDNSTasksRequest) (*pb.FindAllDoingDNSTasksResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	tasks, err := dns.SharedDNSTaskDAO.FindAllDoingOrErrorTasks(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	pbTasks := []*pb.DNSTask{}
	for _, task := range tasks {
		pbTask := &pb.DNSTask{
			Id:        int64(task.Id),
			Type:      task.Type,
			IsDone:    task.IsDone == 1,
			IsOk:      task.IsOk == 1,
			Error:     task.Error,
			UpdatedAt: int64(task.UpdatedAt),
		}

		switch task.Type {
		case dns.DNSTaskTypeClusterChange:
			clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(task.ClusterId))
			if err != nil {
				return nil, err
			}
			if len(clusterName) == 0 {
				clusterName = "集群[" + fmt.Sprintf("%d", task.ClusterId) + "]"
			}
			pbTask.NodeCluster = &pb.NodeCluster{Id: int64(task.ClusterId), Name: clusterName}
		case dns.DNSTaskTypeNodeChange:
			nodeName, err := models.SharedNodeDAO.FindNodeName(tx, int64(task.NodeId))
			if err != nil {
				return nil, err
			}
			if len(nodeName) == 0 {
				nodeName = "节点[" + fmt.Sprintf("%d", task.NodeId) + "]"
			}
			pbTask.Node = &pb.Node{Id: int64(task.NodeId), Name: nodeName}
		case dns.DNSTaskTypeServerChange:
			serverName, err := models.SharedServerDAO.FindEnabledServerName(tx, int64(task.ServerId))
			if err != nil {
				return nil, err
			}
			if len(serverName) == 0 {
				serverName = "服务[" + fmt.Sprintf("%d", task.ServerId) + "]"
			}
			pbTask.Server = &pb.Server{Id: int64(task.ServerId), Name: serverName}
		case dns.DNSTaskTypeDomainChange:
			domainName, err := dns.SharedDNSDomainDAO.FindDNSDomainName(tx, int64(task.DomainId))
			if err != nil {
				return nil, err
			}
			if len(domainName) == 0 {
				domainName = "域名[" + fmt.Sprintf("%d", task.DomainId) + "]"
			}
			pbTask.DnsDomain = &pb.DNSDomain{Id: int64(task.DomainId), Name: domainName}
		}
		pbTasks = append(pbTasks, pbTask)
	}
	return &pb.FindAllDoingDNSTasksResponse{DnsTasks: pbTasks}, nil
}

// DeleteDNSTask 删除任务
func (this *DNSTaskService) DeleteDNSTask(ctx context.Context, req *pb.DeleteDNSTaskRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	err = dns.SharedDNSTaskDAO.DeleteDNSTask(this.NullTx(), req.DnsTaskId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
