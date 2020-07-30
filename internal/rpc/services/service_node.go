package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
)

type NodeService struct {
}

func (this *NodeService) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodeId, err := models.SharedNodeDAO.CreateNode(req.Name, req.ClusterId)
	if err != nil {
		return nil, err
	}

	// 增加认证相关
	if req.Login != nil {
		_, err = models.SharedNodeLoginDAO.CreateNodeLogin(nodeId, req.Login.Name, req.Login.Type, req.Login.Params)
		if err != nil {
			return nil, err
		}
	}

	return &pb.CreateNodeResponse{
		NodeId: nodeId,
	}, nil
}

func (this *NodeService) CountAllEnabledNodes(ctx context.Context, req *pb.CountAllEnabledNodesRequest) (*pb.CountAllEnabledNodesResponse, error) {
	_ = req
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedNodeDAO.CountAllEnabledNodes()
	if err != nil {
		return nil, err
	}

	return &pb.CountAllEnabledNodesResponse{Count: count}, nil
}

func (this *NodeService) ListEnabledNodes(ctx context.Context, req *pb.ListEnabledNodesRequest) (*pb.ListEnabledNodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	nodes, err := models.SharedNodeDAO.ListEnabledNodes(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.Node{}
	for _, node := range nodes {
		// 集群信息
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(node.ClusterId))
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.Node{
			Id:   int64(node.Id),
			Name: node.Name,
			Cluster: &pb.NodeCluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
		})
	}

	return &pb.ListEnabledNodesResponse{
		Nodes: result,
	}, nil
}

// 禁用节点
func (this *NodeService) DisableNode(ctx context.Context, req *pb.DisableNodeRequest) (*pb.DisableNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.DisableNode(req.NodeId)
	if err != nil {
		return nil, err
	}

	return &pb.DisableNodeResponse{}, nil
}

// 修改节点
func (this *NodeService) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.UpdateNode(req.NodeId, req.Name, req.ClusterId)
	if err != nil {
		return nil, err
	}

	if req.Login == nil {
		err = models.SharedNodeLoginDAO.DisableNodeLogins(req.NodeId)
		if err != nil {
			return nil, err
		}
	} else {
		if req.Login.Id > 0 {
			err = models.SharedNodeLoginDAO.UpdateNodeLogin(req.Login.Id, req.Login.Name, req.Login.Type, req.Login.Params)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = models.SharedNodeLoginDAO.CreateNodeLogin(req.NodeId, req.Login.Name, req.Login.Type, req.Login.Params)
			if err != nil {
				return nil, err
			}
		}
	}

	return &pb.UpdateNodeResponse{}, nil
}

// 列出单个节点
func (this *NodeService) FindEnabledNode(ctx context.Context, req *pb.FindEnabledNodeRequest) (*pb.FindEnabledNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	node, err := models.SharedNodeDAO.FindEnabledNode(req.NodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindEnabledNodeResponse{Node: nil}, nil
	}

	// 集群信息
	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(node.ClusterId))
	if err != nil {
		return nil, err
	}

	// 认证信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(req.NodeId)
	if err != nil {
		return nil, err
	}
	var respLogin *pb.NodeLogin = nil
	if login != nil {
		respLogin = &pb.NodeLogin{
			Id:     int64(login.Id),
			Name:   login.Name,
			Type:   login.Type,
			Params: []byte(login.Params),
		}
	}

	return &pb.FindEnabledNodeResponse{Node: &pb.Node{
		Id:   int64(node.Id),
		Name: node.Name,
		Cluster: &pb.NodeCluster{
			Id:   int64(node.ClusterId),
			Name: clusterName,
		},
		Login: respLogin,
	}}, nil
}
