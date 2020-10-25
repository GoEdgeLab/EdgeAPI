package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/installers"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
)

// 边缘节点相关服务
type NodeService struct {
}

// 创建节点
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

// 注册集群节点
func (this *NodeService) RegisterClusterNode(ctx context.Context, req *pb.RegisterClusterNodeRequest) (*pb.RegisterClusterNodeResponse, error) {
	// 校验请求
	_, clusterId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeCluster)
	if err != nil {
		return nil, err
	}

	nodeId, err := models.SharedNodeDAO.CreateNode(req.Name, clusterId)
	if err != nil {
		return nil, err
	}
	err = models.SharedNodeDAO.UpdateNodeIsInstalled(nodeId, true)
	if err != nil {
		return nil, err
	}

	node, err := models.SharedNodeDAO.FindEnabledNode(nodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("can not find node after creating")
	}

	// 获取集群可以使用的所有API节点
	apiAddrs, err := models.SharedNodeClusterDAO.FindAllAPINodeAddrsWithCluster(clusterId)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterClusterNodeResponse{
		UniqueId:  node.UniqueId,
		Secret:    node.Secret,
		Endpoints: apiAddrs,
	}, nil
}

// 计算节点数量
func (this *NodeService) CountAllEnabledNodes(ctx context.Context, req *pb.CountAllEnabledNodesRequest) (*pb.CountAllEnabledNodesResponse, error) {
	// 校验请求
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
	count, err := models.SharedNodeDAO.CountAllEnabledNodesMatch(req.ClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState))
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledNodesMatchResponse{Count: count}, nil
}

// 列出单页的节点
func (this *NodeService) ListEnabledNodesMatch(ctx context.Context, req *pb.ListEnabledNodesMatchRequest) (*pb.ListEnabledNodesMatchResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	nodes, err := models.SharedNodeDAO.ListEnabledNodesMatch(req.Offset, req.Size, req.ClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState))
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

		// 安装信息
		installStatus, err := node.DecodeInstallStatus()
		if err != nil {
			return nil, err
		}
		installStatusResult := &pb.NodeInstallStatus{}
		if installStatus != nil {
			installStatusResult = &pb.NodeInstallStatus{
				IsRunning:  installStatus.IsRunning,
				IsFinished: installStatus.IsFinished,
				IsOk:       installStatus.IsOk,
				Error:      installStatus.Error,
				UpdatedAt:  installStatus.UpdatedAt,
			}
		}

		result = append(result, &pb.Node{
			Id:          int64(node.Id),
			Name:        node.Name,
			Version:     int64(node.Version),
			IsInstalled: node.IsInstalled == 1,
			Status:      node.Status,
			Cluster: &pb.NodeCluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
			InstallStatus: installStatusResult,
			MaxCPU:        types.Int32(node.MaxCPU),
			IsOn:          node.IsOn == 1,
		})
	}

	return &pb.ListEnabledNodesMatchResponse{
		Nodes: result,
	}, nil
}

// 查找一个集群下的所有节点
func (this *NodeService) FindAllEnabledNodesWithClusterId(ctx context.Context, req *pb.FindAllEnabledNodesWithClusterIdRequest) (*pb.FindAllEnabledNodesWithClusterIdResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesWithClusterId(req.ClusterId)
	if err != nil {
		return nil, err
	}
	result := []*pb.Node{}
	for _, node := range nodes {
		apiNodeIds := []int64{}
		if models.IsNotNull(node.ConnectedAPINodes) {
			err = json.Unmarshal([]byte(node.ConnectedAPINodes), &apiNodeIds)
			if err != nil {
				return nil, err
			}
		}

		result = append(result, &pb.Node{
			Id:                  int64(node.Id),
			Name:                node.Name,
			UniqueId:            node.UniqueId,
			Secret:              node.Secret,
			ConnectedAPINodeIds: apiNodeIds,
			MaxCPU:              types.Int32(node.MaxCPU),
			IsOn:                node.IsOn == 1,
		})
	}
	return &pb.FindAllEnabledNodesWithClusterIdResponse{Nodes: result}, nil
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
func (this *NodeService) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.UpdateNode(req.NodeId, req.Name, req.ClusterId, req.MaxCPU, req.IsOn)
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

	return &pb.RPCUpdateSuccess{}, nil
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

	// 安装信息
	installStatus, err := node.DecodeInstallStatus()
	if err != nil {
		return nil, err
	}
	installStatusResult := &pb.NodeInstallStatus{}
	if installStatus != nil {
		installStatusResult = &pb.NodeInstallStatus{
			IsRunning:  installStatus.IsRunning,
			IsFinished: installStatus.IsFinished,
			IsOk:       installStatus.IsOk,
			Error:      installStatus.Error,
			UpdatedAt:  installStatus.UpdatedAt,
		}
	}

	return &pb.FindEnabledNodeResponse{Node: &pb.Node{
		Id:            int64(node.Id),
		Name:          node.Name,
		Status:        node.Status,
		UniqueId:      node.UniqueId,
		Version:       int64(node.Version),
		LatestVersion: int64(node.LatestVersion),
		Secret:        node.Secret,
		InstallDir:    node.InstallDir,
		IsInstalled:   node.IsInstalled == 1,
		Cluster: &pb.NodeCluster{
			Id:   int64(node.ClusterId),
			Name: clusterName,
		},
		Login:         respLogin,
		InstallStatus: installStatusResult,
		MaxCPU:        types.Int32(node.MaxCPU),
		IsOn:          node.IsOn == 1,
	}}, nil
}

// 组合节点配置
func (this *NodeService) ComposeNodeConfig(ctx context.Context, req *pb.ComposeNodeConfigRequest) (*pb.ComposeNodeConfigResponse, error) {
	_ = req

	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}
	nodeConfig, err := models.SharedNodeDAO.ComposeNodeConfig(nodeId)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(nodeConfig)
	if err != nil {
		return nil, err
	}

	return &pb.ComposeNodeConfigResponse{NodeJSON: data}, nil
}

// 更新节点状态
func (this *NodeService) UpdateNodeStatus(ctx context.Context, req *pb.UpdateNodeStatusRequest) (*pb.RPCUpdateSuccess, error) {
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

	return &pb.RPCUpdateSuccess{}, nil
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

// 修改节点安装状态
func (this *NodeService) UpdateNodeIsInstalled(ctx context.Context, req *pb.UpdateNodeIsInstalledRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.UpdateNodeIsInstalled(req.NodeId, req.IsInstalled)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 安装节点
func (this *NodeService) InstallNode(ctx context.Context, req *pb.InstallNodeRequest) (*pb.InstallNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	go func() {
		err = installers.SharedQueue().InstallNodeProcess(req.NodeId)
		if err != nil {
			logs.Println("[RPC]install node:" + err.Error())
		}
	}()

	return &pb.InstallNodeResponse{}, nil
}

// 更改节点连接的API节点信息
func (this *NodeService) UpdateNodeConnectedAPINodes(ctx context.Context, req *pb.UpdateNodeConnectedAPINodesRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.UpdateNodeConnectedAPINodes(nodeId, req.ApiNodeIds)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return rpcutils.RPCUpdateSuccess()
}

// 计算使用某个认证的节点数量
func (this *NodeService) CountAllEnabledNodesWithGrantId(ctx context.Context, req *pb.CountAllEnabledNodesWithGrantIdRequest) (*pb.CountAllEnabledNodesWithGrantIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedNodeDAO.CountAllEnabledNodesWithGrantId(req.GrantId)
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledNodesWithGrantIdResponse{Count: count}, nil
}

// 查找使用某个认证的所有节点
func (this *NodeService) FindAllEnabledNodesWithGrantId(ctx context.Context, req *pb.FindAllEnabledNodesWithGrantIdRequest) (*pb.FindAllEnabledNodesWithGrantIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesWithGrantId(req.GrantId)
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
			Id:          int64(node.Id),
			Name:        node.Name,
			Version:     int64(node.Version),
			IsInstalled: node.IsInstalled == 1,
			Status:      node.Status,
			Cluster: &pb.NodeCluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
			IsOn: node.IsOn == 1,
		})
	}

	return &pb.FindAllEnabledNodesWithGrantIdResponse{Nodes: result}, nil
}
