package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
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

// 计算匹配的节点数量
func (this *NodeService) CountAllEnabledNodesMatch(ctx context.Context, req *pb.CountAllEnabledNodesMatchRequest) (*pb.CountAllEnabledNodesMatchResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedNodeDAO.CountAllEnabledNodesMatch(req.ClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledNodesMatchResponse{Count: count}, nil
}

func (this *NodeService) ListEnabledNodesMatch(ctx context.Context, req *pb.ListEnabledNodesMatchRequest) (*pb.ListEnabledNodesMatchResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	nodes, err := models.SharedNodeDAO.ListEnabledNodesMatch(req.Offset, req.Size, req.ClusterId)
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
			Id:     int64(node.Id),
			Name:   node.Name,
			Status: node.Status,
			Cluster: &pb.NodeCluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
		})
	}

	return &pb.ListEnabledNodesMatchResponse{
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
		Id:          int64(node.Id),
		Name:        node.Name,
		Status:      node.Status,
		UniqueId:    node.UniqueId,
		Secret:      node.Secret,
		InstallDir:  node.InstallDir,
		IsInstalled: node.IsInstalled == 1,
		Cluster: &pb.NodeCluster{
			Id:   int64(node.ClusterId),
			Name: clusterName,
		},
		Login: respLogin,
	}}, nil
}

// 组合节点配置
func (this *NodeService) ComposeNodeConfig(ctx context.Context, req *pb.ComposeNodeConfigRequest) (*pb.ComposeNodeConfigResponse, error) {
	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	node, err := models.SharedNodeDAO.FindEnabledNode(nodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node validate failed, please check 'nodeId' or 'secret'")
	}

	nodeMap := maps.Map{
		"id":      node.UniqueId,
		"isOn":    node.IsOn == 1,
		"servers": []maps.Map{},
		"version": node.Version,
	}

	// 获取所有的服务
	servers, err := models.SharedServerDAO.FindAllEnabledServersWithNode(int64(node.Id))
	if err != nil {
		return nil, err
	}

	serverMaps := []maps.Map{}
	for _, server := range servers {
		if len(server.Config) == 0 {
			continue
		}
		configMap := maps.Map{}
		err = json.Unmarshal([]byte(server.Config), &configMap)
		if err != nil {
			return nil, err
		}
		configMap["id"] = server.UniqueId
		configMap["version"] = server.Version
		serverMaps = append(serverMaps, configMap)
	}
	nodeMap["servers"] = serverMaps

	data, err := json.Marshal(nodeMap)
	if err != nil {
		return nil, err
	}

	return &pb.ComposeNodeConfigResponse{ConfigJSON: data}, nil
}

// 节点stream
func (this *NodeService) NodeStream(server pb.NodeService_NodeStreamServer) error {
	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(server.Context(), rpcutils.UserTypeNode)
	if err != nil {
		return err
	}
	logs.Println("nodeId:", nodeId)

	for {
		req, err := server.Recv()
		if err != nil {
			return err
		}
		logs.Println("received:", req)
	}
}

// 更新节点状态
func (this *NodeService) UpdateNodeStatus(ctx context.Context, req *pb.UpdateNodeStatusRequest) (*pb.UpdateNodeStatusResponse, error) {
	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	if req.NodeId > 0 {
		nodeId = req.NodeId
	}

	if nodeId <= 0 {
		return nil, errors.New("'nodeId' should be greater than 0")
	}

	err = models.SharedNodeDAO.UpdateNodeStatus(nodeId, req.StatusJSON)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateNodeStatusResponse{}, nil
}

// 同步集群中的节点版本
func (this *NodeService) SyncNodesVersionWithCluster(ctx context.Context, req *pb.SyncNodesVersionWithClusterRequest) (*pb.SyncNodesVersionWithClusterResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.SyncNodeVersionsWithCluster(req.ClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.SyncNodesVersionWithClusterResponse{}, nil
}
