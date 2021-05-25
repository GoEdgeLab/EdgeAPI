package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/installers"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"net"
)

// NodeService 边缘节点相关服务
type NodeService struct {
	BaseService
}

// CreateNode 创建节点
func (this *NodeService) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodeId, err := models.SharedNodeDAO.CreateNode(tx, adminId, req.Name, req.NodeClusterId, req.NodeGroupId, req.NodeRegionId)
	if err != nil {
		return nil, err
	}

	// 增加认证相关
	if req.NodeLogin != nil {
		_, err = models.SharedNodeLoginDAO.CreateNodeLogin(tx, nodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
		if err != nil {
			return nil, err
		}
	}

	// 保存DNS相关
	if req.DnsDomainId > 0 && len(req.DnsRoutes) > 0 {
		err = models.SharedNodeDAO.UpdateNodeDNS(tx, nodeId, map[int64][]string{
			req.DnsDomainId: req.DnsRoutes,
		})
		if err != nil {
			return nil, err
		}
	}

	return &pb.CreateNodeResponse{
		NodeId: nodeId,
	}, nil
}

// RegisterClusterNode 注册集群节点
func (this *NodeService) RegisterClusterNode(ctx context.Context, req *pb.RegisterClusterNodeRequest) (*pb.RegisterClusterNodeResponse, error) {
	// 校验请求
	_, clusterId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeCluster)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	adminId, err := models.SharedNodeClusterDAO.FindClusterAdminId(tx, clusterId)
	if err != nil {
		return nil, err
	}

	nodeId, err := models.SharedNodeDAO.CreateNode(tx, adminId, req.Name, clusterId, 0, 0)
	if err != nil {
		return nil, err
	}
	err = models.SharedNodeDAO.UpdateNodeIsInstalled(tx, nodeId, true)
	if err != nil {
		return nil, err
	}

	node, err := models.SharedNodeDAO.FindEnabledNode(tx, nodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("can not find node after creating")
	}

	// 获取集群可以使用的所有API节点
	apiAddrs, err := models.SharedNodeClusterDAO.FindAllAPINodeAddrsWithCluster(tx, clusterId)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterClusterNodeResponse{
		UniqueId:  node.UniqueId,
		Secret:    node.Secret,
		Endpoints: apiAddrs,
	}, nil
}

// CountAllEnabledNodes 计算节点数量
func (this *NodeService) CountAllEnabledNodes(ctx context.Context, req *pb.CountAllEnabledNodesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// CountAllEnabledNodesMatch 计算匹配的节点数量
func (this *NodeService) CountAllEnabledNodesMatch(ctx context.Context, req *pb.CountAllEnabledNodesMatchRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeDAO.CountAllEnabledNodesMatch(tx, req.NodeClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState), req.Keyword, req.NodeGroupId, req.NodeRegionId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledNodesMatch 列出单页的节点
func (this *NodeService) ListEnabledNodesMatch(ctx context.Context, req *pb.ListEnabledNodesMatchRequest) (*pb.ListEnabledNodesMatchResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	dnsDomainId := int64(0)
	domainRoutes := []*dnsclients.Route{}
	if clusterDNS != nil {
		dnsDomainId = int64(clusterDNS.DnsDomainId)
		if clusterDNS.DnsDomainId > 0 {
			domainRoutes, err = dns.SharedDNSDomainDAO.FindDomainRoutes(tx, dnsDomainId)
			if err != nil {
				return nil, err
			}
		}
	}

	nodes, err := models.SharedNodeDAO.ListEnabledNodesMatch(tx, req.Offset, req.Size, req.NodeClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState), req.Keyword, req.NodeGroupId, req.NodeRegionId)
	if err != nil {
		return nil, err
	}
	result := []*pb.Node{}
	for _, node := range nodes {
		// 集群信息
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(node.ClusterId))
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
				ErrorCode:  installStatus.ErrorCode,
				UpdatedAt:  installStatus.UpdatedAt,
			}
		}

		// 分组信息
		var pbGroup *pb.NodeGroup = nil
		if node.GroupId > 0 {
			group, err := models.SharedNodeGroupDAO.FindEnabledNodeGroup(tx, int64(node.GroupId))
			if err != nil {
				return nil, err
			}
			if group != nil {
				pbGroup = &pb.NodeGroup{
					Id:   int64(group.Id),
					Name: group.Name,
				}
			}
		}

		// DNS线路
		routeCodes, err := node.DNSRouteCodesForDomainId(dnsDomainId)
		if err != nil {
			return nil, err
		}
		pbRoutes := []*pb.DNSRoute{}
		for _, routeCode := range routeCodes {
			for _, route := range domainRoutes {
				if route.Code == routeCode {
					pbRoutes = append(pbRoutes, &pb.DNSRoute{
						Name: route.Name,
						Code: route.Code,
					})
					break
				}
			}
		}

		// 区域
		var pbRegion *pb.NodeRegion = nil
		if node.RegionId > 0 {
			region, err := models.SharedNodeRegionDAO.FindEnabledNodeRegion(tx, int64(node.RegionId))
			if err != nil {
				return nil, err
			}
			if region != nil {
				pbRegion = &pb.NodeRegion{
					Id:   int64(region.Id),
					IsOn: region.IsOn == 1,
					Name: region.Name,
				}
			}
		}

		result = append(result, &pb.Node{
			Id:          int64(node.Id),
			Name:        node.Name,
			Version:     int64(node.Version),
			IsInstalled: node.IsInstalled == 1,
			StatusJSON:  []byte(node.Status),
			NodeCluster: &pb.NodeCluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
			InstallStatus: installStatusResult,
			MaxCPU:        types.Int32(node.MaxCPU),
			IsOn:          node.IsOn == 1,
			IsUp:          node.IsUp == 1,
			Group:         pbGroup,
			Region:        pbRegion,
			DnsRoutes:     pbRoutes,
		})
	}

	return &pb.ListEnabledNodesMatchResponse{
		Nodes: result,
	}, nil
}

// FindAllEnabledNodesWithNodeClusterId 查找一个集群下的所有节点
func (this *NodeService) FindAllEnabledNodesWithNodeClusterId(ctx context.Context, req *pb.FindAllEnabledNodesWithNodeClusterIdRequest) (*pb.FindAllEnabledNodesWithNodeClusterIdResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查权限
	}

	tx := this.NullTx()

	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesWithClusterId(tx, req.NodeClusterId)
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
	return &pb.FindAllEnabledNodesWithNodeClusterIdResponse{Nodes: result}, nil
}

// DeleteNode 删除节点
func (this *NodeService) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeDAO.DisableNode(tx, req.NodeId)
	if err != nil {
		return nil, err
	}

	// 删除节点相关任务
	err = models.SharedNodeTaskDAO.DeleteNodeTasks(tx, req.NodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNode 修改节点
func (this *NodeService) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	var maxCacheDiskCapacityJSON []byte
	if req.MaxCacheDiskCapacity != nil {
		maxCacheDiskCapacityJSON, err = json.Marshal(&shared.SizeCapacity{
			Count: req.MaxCacheDiskCapacity.Count,
			Unit:  req.MaxCacheDiskCapacity.Unit,
		})
		if err != nil {
			return nil, err
		}
	}

	var maxCacheMemoryCapacityJSON []byte
	if req.MaxCacheMemoryCapacity != nil {
		maxCacheMemoryCapacityJSON, err = json.Marshal(&shared.SizeCapacity{
			Count: req.MaxCacheMemoryCapacity.Count,
			Unit:  req.MaxCacheMemoryCapacity.Unit,
		})
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedNodeDAO.UpdateNode(tx, req.NodeId, req.Name, req.NodeClusterId, req.NodeGroupId, req.NodeRegionId, req.MaxCPU, req.IsOn, maxCacheDiskCapacityJSON, maxCacheMemoryCapacityJSON)
	if err != nil {
		return nil, err
	}

	if req.NodeLogin == nil {
		err = models.SharedNodeLoginDAO.DisableNodeLogins(tx, req.NodeId)
		if err != nil {
			return nil, err
		}
	} else {
		if req.NodeLogin.Id > 0 {
			err = models.SharedNodeLoginDAO.UpdateNodeLogin(tx, req.NodeLogin.Id, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = models.SharedNodeLoginDAO.CreateNodeLogin(tx, req.NodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
			if err != nil {
				return nil, err
			}
		}
	}

	// 保存DNS相关
	if req.DnsDomainId > 0 && len(req.DnsRoutes) > 0 {
		err = models.SharedNodeDAO.UpdateNodeDNS(tx, req.NodeId, map[int64][]string{
			req.DnsDomainId: req.DnsRoutes,
		})
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// FindEnabledNode 列出单个节点
func (this *NodeService) FindEnabledNode(ctx context.Context, req *pb.FindEnabledNodeRequest) (*pb.FindEnabledNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedNodeDAO.FindEnabledNode(tx, req.NodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindEnabledNodeResponse{Node: nil}, nil
	}

	// 集群信息
	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(node.ClusterId))
	if err != nil {
		return nil, err
	}

	// 认证信息
	login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(tx, req.NodeId)
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
			ErrorCode:  installStatus.ErrorCode,
			UpdatedAt:  installStatus.UpdatedAt,
		}
	}

	// 分组信息
	var pbGroup *pb.NodeGroup = nil
	if node.GroupId > 0 {
		group, err := models.SharedNodeGroupDAO.FindEnabledNodeGroup(tx, int64(node.GroupId))
		if err != nil {
			return nil, err
		}
		if group != nil {
			pbGroup = &pb.NodeGroup{
				Id:   int64(group.Id),
				Name: group.Name,
			}
		}
	}

	// 区域
	var pbRegion *pb.NodeRegion = nil
	if node.RegionId > 0 {
		region, err := models.SharedNodeRegionDAO.FindEnabledNodeRegion(tx, int64(node.RegionId))
		if err != nil {
			return nil, err
		}
		if region != nil {
			pbRegion = &pb.NodeRegion{
				Id:   int64(region.Id),
				IsOn: region.IsOn == 1,
				Name: region.Name,
			}
		}
	}

	// 最大硬盘容量
	var pbMaxCacheDiskCapacity *pb.SizeCapacity
	if len(node.MaxCacheDiskCapacity) > 0 {
		pbMaxCacheDiskCapacity = &pb.SizeCapacity{}
		err = json.Unmarshal([]byte(node.MaxCacheDiskCapacity), pbMaxCacheDiskCapacity)
		if err != nil {
			return nil, err
		}
	}

	// 最大内存容量
	var pbMaxCacheMemoryCapacity *pb.SizeCapacity
	if len(node.MaxCacheMemoryCapacity) > 0 {
		pbMaxCacheMemoryCapacity = &pb.SizeCapacity{}
		err = json.Unmarshal([]byte(node.MaxCacheMemoryCapacity), pbMaxCacheMemoryCapacity)
		if err != nil {
			return nil, err
		}
	}

	return &pb.FindEnabledNodeResponse{Node: &pb.Node{
		Id:            int64(node.Id),
		Name:          node.Name,
		StatusJSON:    []byte(node.Status),
		UniqueId:      node.UniqueId,
		Version:       int64(node.Version),
		LatestVersion: int64(node.LatestVersion),
		Secret:        node.Secret,
		InstallDir:    node.InstallDir,
		IsInstalled:   node.IsInstalled == 1,
		NodeCluster: &pb.NodeCluster{
			Id:   int64(node.ClusterId),
			Name: clusterName,
		},
		Login:                  respLogin,
		InstallStatus:          installStatusResult,
		MaxCPU:                 types.Int32(node.MaxCPU),
		IsOn:                   node.IsOn == 1,
		Group:                  pbGroup,
		Region:                 pbRegion,
		MaxCacheDiskCapacity:   pbMaxCacheDiskCapacity,
		MaxCacheMemoryCapacity: pbMaxCacheMemoryCapacity,
	}}, nil
}

// FindCurrentNodeConfig 组合节点配置
func (this *NodeService) FindCurrentNodeConfig(ctx context.Context, req *pb.FindCurrentNodeConfigRequest) (*pb.FindCurrentNodeConfigResponse, error) {
	_ = req

	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查版本号
	currentVersion, err := models.SharedNodeDAO.FindNodeVersion(tx, nodeId)
	if err != nil {
		return nil, err
	}
	if currentVersion == req.Version {
		return &pb.FindCurrentNodeConfigResponse{IsChanged: false}, nil
	}

	nodeConfig, err := models.SharedNodeDAO.ComposeNodeConfig(tx, nodeId)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(nodeConfig)
	if err != nil {
		return nil, err
	}

	return &pb.FindCurrentNodeConfigResponse{IsChanged: true, NodeJSON: data}, nil
}

// UpdateNodeStatus 更新节点状态
func (this *NodeService) UpdateNodeStatus(ctx context.Context, req *pb.UpdateNodeStatusRequest) (*pb.RPCSuccess, error) {
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

	tx := this.NullTx()

	err = models.SharedNodeDAO.UpdateNodeStatus(tx, nodeId, req.StatusJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateNodeIsInstalled 修改节点安装状态
func (this *NodeService) UpdateNodeIsInstalled(ctx context.Context, req *pb.UpdateNodeIsInstalledRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeDAO.UpdateNodeIsInstalled(tx, req.NodeId, req.IsInstalled)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// InstallNode 安装节点
func (this *NodeService) InstallNode(ctx context.Context, req *pb.InstallNodeRequest) (*pb.InstallNodeResponse, error) {
	// 校验节点
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	go func() {
		err = installers.SharedQueue().InstallNodeProcess(req.NodeId, false)
		if err != nil {
			logs.Println("[RPC]install node:" + err.Error())
		}
	}()

	return &pb.InstallNodeResponse{}, nil
}

// UpgradeNode 升级节点
func (this *NodeService) UpgradeNode(ctx context.Context, req *pb.UpgradeNodeRequest) (*pb.UpgradeNodeResponse, error) {
	// 校验节点
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeDAO.UpdateNodeIsInstalled(tx, req.NodeId, false)
	if err != nil {
		return nil, err
	}

	// 检查状态
	installStatus, err := models.SharedNodeDAO.FindNodeInstallStatus(tx, req.NodeId)
	if err != nil {
		return nil, err
	}
	if installStatus == nil {
		installStatus = &models.NodeInstallStatus{}
	}
	installStatus.IsOk = false
	installStatus.IsFinished = false
	err = models.SharedNodeDAO.UpdateNodeInstallStatus(tx, req.NodeId, installStatus)
	if err != nil {
		return nil, err
	}

	go func() {
		err = installers.SharedQueue().InstallNodeProcess(req.NodeId, true)
		if err != nil {
			logs.Println("[RPC]install node:" + err.Error())
		}
	}()

	return &pb.UpgradeNodeResponse{}, nil
}

// StartNode 启动节点
func (this *NodeService) StartNode(ctx context.Context, req *pb.StartNodeRequest) (*pb.StartNodeResponse, error) {
	// 校验节点
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = installers.SharedQueue().StartNode(req.NodeId)
	if err != nil {
		return &pb.StartNodeResponse{
			IsOk:  false,
			Error: err.Error(),
		}, nil
	}

	return &pb.StartNodeResponse{IsOk: true}, nil
}

// StopNode 停止节点
func (this *NodeService) StopNode(ctx context.Context, req *pb.StopNodeRequest) (*pb.StopNodeResponse, error) {
	// 校验节点
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = installers.SharedQueue().StopNode(req.NodeId)
	if err != nil {
		return &pb.StopNodeResponse{
			IsOk:  false,
			Error: err.Error(),
		}, nil
	}

	return &pb.StopNodeResponse{IsOk: true}, nil
}

// UpdateNodeConnectedAPINodes 更改节点连接的API节点信息
func (this *NodeService) UpdateNodeConnectedAPINodes(ctx context.Context, req *pb.UpdateNodeConnectedAPINodesRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeDAO.UpdateNodeConnectedAPINodes(tx, nodeId, req.ApiNodeIds)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return this.Success()
}

// CountAllEnabledNodesWithNodeGrantId 计算使用某个认证的节点数量
func (this *NodeService) CountAllEnabledNodesWithNodeGrantId(ctx context.Context, req *pb.CountAllEnabledNodesWithNodeGrantIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeDAO.CountAllEnabledNodesWithGrantId(tx, req.GrantId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllEnabledNodesWithNodeGrantId 查找使用某个认证的所有节点
func (this *NodeService) FindAllEnabledNodesWithNodeGrantId(ctx context.Context, req *pb.FindAllEnabledNodesWithNodeGrantIdRequest) (*pb.FindAllEnabledNodesWithNodeGrantIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesWithGrantId(tx, req.GrantId)
	if err != nil {
		return nil, err
	}

	result := []*pb.Node{}
	for _, node := range nodes {
		// 集群信息
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(node.ClusterId))
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.Node{
			Id:          int64(node.Id),
			Name:        node.Name,
			Version:     int64(node.Version),
			IsInstalled: node.IsInstalled == 1,
			StatusJSON:  []byte(node.Status),
			NodeCluster: &pb.NodeCluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
			IsOn: node.IsOn == 1,
		})
	}

	return &pb.FindAllEnabledNodesWithNodeGrantIdResponse{Nodes: result}, nil
}

// CountAllNotInstalledNodesWithNodeClusterId 计算没有安装的节点数量
func (this *NodeService) CountAllNotInstalledNodesWithNodeClusterId(ctx context.Context, req *pb.CountAllNotInstalledNodesWithNodeClusterIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	tx := this.NullTx()
	count, err := models.SharedNodeDAO.CountAllNotInstalledNodesWithClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllNotInstalledNodesWithNodeClusterId 列出所有未安装的节点
func (this *NodeService) FindAllNotInstalledNodesWithNodeClusterId(ctx context.Context, req *pb.FindAllNotInstalledNodesWithNodeClusterIdRequest) (*pb.FindAllNotInstalledNodesWithNodeClusterIdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedNodeDAO.FindAllNotInstalledNodesWithClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result := []*pb.Node{}
	for _, node := range nodes {
		// 认证信息
		login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(tx, int64(node.Id))
		if err != nil {
			return nil, err
		}
		var pbLogin *pb.NodeLogin = nil
		if login != nil {
			pbLogin = &pb.NodeLogin{
				Id:     int64(login.Id),
				Name:   login.Name,
				Type:   login.Type,
				Params: []byte(login.Params),
			}
		}

		// IP信息
		addresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(tx, int64(node.Id))
		if err != nil {
			return nil, err
		}

		pbAddresses := []*pb.NodeIPAddress{}
		for _, address := range addresses {
			pbAddresses = append(pbAddresses, &pb.NodeIPAddress{
				Id:          int64(address.Id),
				NodeId:      int64(address.NodeId),
				Name:        address.Name,
				Ip:          address.Ip,
				Description: address.Description,
				State:       int64(address.State),
				Order:       int64(address.Order),
				CanAccess:   address.CanAccess == 1,
			})
		}

		// 安装信息
		installStatus, err := node.DecodeInstallStatus()
		if err != nil {
			return nil, err
		}
		pbInstallStatus := &pb.NodeInstallStatus{}
		if installStatus != nil {
			pbInstallStatus = &pb.NodeInstallStatus{
				IsRunning:  installStatus.IsRunning,
				IsFinished: installStatus.IsFinished,
				IsOk:       installStatus.IsOk,
				Error:      installStatus.Error,
				ErrorCode:  installStatus.ErrorCode,
				UpdatedAt:  installStatus.UpdatedAt,
			}
		}

		result = append(result, &pb.Node{
			Id:            int64(node.Id),
			Name:          node.Name,
			Version:       int64(node.Version),
			IsInstalled:   node.IsInstalled == 1,
			StatusJSON:    []byte(node.Status),
			IsOn:          node.IsOn == 1,
			Login:         pbLogin,
			IpAddresses:   pbAddresses,
			InstallStatus: pbInstallStatus,
		})
	}
	return &pb.FindAllNotInstalledNodesWithNodeClusterIdResponse{Nodes: result}, nil
}

// CountAllUpgradeNodesWithNodeClusterId 计算需要升级的节点数量
func (this *NodeService) CountAllUpgradeNodesWithNodeClusterId(ctx context.Context, req *pb.CountAllUpgradeNodesWithNodeClusterIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	deployFiles := installers.SharedDeployManager.LoadFiles()
	total := int64(0)
	for _, deployFile := range deployFiles {
		count, err := models.SharedNodeDAO.CountAllLowerVersionNodesWithClusterId(tx, req.NodeClusterId, deployFile.OS, deployFile.Arch, deployFile.Version)
		if err != nil {
			return nil, err
		}
		total += count
	}

	return this.SuccessCount(total)
}

// FindAllUpgradeNodesWithNodeClusterId 列出所有需要升级的节点
func (this *NodeService) FindAllUpgradeNodesWithNodeClusterId(ctx context.Context, req *pb.FindAllUpgradeNodesWithNodeClusterIdRequest) (*pb.FindAllUpgradeNodesWithNodeClusterIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 获取当前能升级到的最新版本
	deployFiles := installers.SharedDeployManager.LoadFiles()
	result := []*pb.FindAllUpgradeNodesWithNodeClusterIdResponse_NodeUpgrade{}
	for _, deployFile := range deployFiles {
		nodes, err := models.SharedNodeDAO.FindAllLowerVersionNodesWithClusterId(tx, req.NodeClusterId, deployFile.OS, deployFile.Arch, deployFile.Version)
		if err != nil {
			return nil, err
		}
		for _, node := range nodes {
			// 认证信息
			login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(tx, int64(node.Id))
			if err != nil {
				return nil, err
			}
			var pbLogin *pb.NodeLogin = nil
			if login != nil {
				pbLogin = &pb.NodeLogin{
					Id:     int64(login.Id),
					Name:   login.Name,
					Type:   login.Type,
					Params: []byte(login.Params),
				}
			}

			// IP信息
			addresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(tx, int64(node.Id))
			if err != nil {
				return nil, err
			}

			pbAddresses := []*pb.NodeIPAddress{}
			for _, address := range addresses {
				pbAddresses = append(pbAddresses, &pb.NodeIPAddress{
					Id:          int64(address.Id),
					NodeId:      int64(address.NodeId),
					Name:        address.Name,
					Ip:          address.Ip,
					Description: address.Description,
					State:       int64(address.State),
					Order:       int64(address.Order),
					CanAccess:   address.CanAccess == 1,
				})
			}

			// 状态
			status, err := node.DecodeStatus()
			if err != nil {
				return nil, err
			}
			if status == nil {
				continue
			}

			// 安装信息
			installStatus, err := node.DecodeInstallStatus()
			if err != nil {
				return nil, err
			}
			pbInstallStatus := &pb.NodeInstallStatus{}
			if installStatus != nil {
				pbInstallStatus = &pb.NodeInstallStatus{
					IsRunning:  installStatus.IsRunning,
					IsFinished: installStatus.IsFinished,
					IsOk:       installStatus.IsOk,
					Error:      installStatus.Error,
					ErrorCode:  installStatus.ErrorCode,
					UpdatedAt:  installStatus.UpdatedAt,
				}
			}

			pbNode := &pb.Node{
				Id:            int64(node.Id),
				Name:          node.Name,
				Version:       int64(node.Version),
				IsInstalled:   node.IsInstalled == 1,
				StatusJSON:    []byte(node.Status),
				IsOn:          node.IsOn == 1,
				IpAddresses:   pbAddresses,
				Login:         pbLogin,
				InstallStatus: pbInstallStatus,
			}

			result = append(result, &pb.FindAllUpgradeNodesWithNodeClusterIdResponse_NodeUpgrade{
				Os:         deployFile.OS,
				Arch:       deployFile.Arch,
				OldVersion: status.BuildVersion,
				NewVersion: deployFile.Version,
				Node:       pbNode,
			})
		}
	}
	return &pb.FindAllUpgradeNodesWithNodeClusterIdResponse{
		Nodes: result,
	}, nil
}

// FindNodeInstallStatus 读取节点安装状态
func (this *NodeService) FindNodeInstallStatus(ctx context.Context, req *pb.FindNodeInstallStatusRequest) (*pb.FindNodeInstallStatusResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	installStatus, err := models.SharedNodeDAO.FindNodeInstallStatus(tx, req.NodeId)
	if err != nil {
		return nil, err
	}
	if installStatus == nil {
		return &pb.FindNodeInstallStatusResponse{InstallStatus: nil}, nil
	}

	pbInstallStatus := &pb.NodeInstallStatus{
		IsRunning:  installStatus.IsRunning,
		IsFinished: installStatus.IsFinished,
		IsOk:       installStatus.IsOk,
		Error:      installStatus.Error,
		ErrorCode:  installStatus.ErrorCode,
		UpdatedAt:  installStatus.UpdatedAt,
	}
	return &pb.FindNodeInstallStatusResponse{InstallStatus: pbInstallStatus}, nil
}

// UpdateNodeLogin 修改节点登录信息
func (this *NodeService) UpdateNodeLogin(ctx context.Context, req *pb.UpdateNodeLoginRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if req.NodeLogin.Id <= 0 {
		_, err := models.SharedNodeLoginDAO.CreateNodeLogin(tx, req.NodeId, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedNodeLoginDAO.UpdateNodeLogin(tx, req.NodeLogin.Id, req.NodeLogin.Name, req.NodeLogin.Type, req.NodeLogin.Params)

	return this.Success()
}

// CountAllEnabledNodesWithNodeGroupId 计算某个节点分组内的节点数量
func (this *NodeService) CountAllEnabledNodesWithNodeGroupId(ctx context.Context, req *pb.CountAllEnabledNodesWithNodeGroupIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeDAO.CountAllEnabledNodesWithGroupId(tx, req.NodeGroupId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllEnabledNodesDNSWithNodeClusterId 取得某个集群下的所有节点
func (this *NodeService) FindAllEnabledNodesDNSWithNodeClusterId(ctx context.Context, req *pb.FindAllEnabledNodesDNSWithNodeClusterIdRequest) (*pb.FindAllEnabledNodesDNSWithNodeClusterIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	if clusterDNS == nil {
		return nil, errors.New("not found clusterId '" + numberutils.FormatInt64(req.NodeClusterId) + "'")
	}
	dnsDomainId := int64(clusterDNS.DnsDomainId)

	routes, err := dns.SharedDNSDomainDAO.FindDomainRoutes(tx, dnsDomainId)
	if err != nil {
		return nil, err
	}

	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeDNSInfo{}
	for _, node := range nodes {
		ipAddresses, err := models.SharedNodeIPAddressDAO.FindNodeAccessIPAddresses(tx, int64(node.Id))
		if err != nil {
			return nil, err
		}

		domainRouteCodes, err := node.DNSRouteCodesForDomainId(dnsDomainId)
		if err != nil {
			return nil, err
		}

		pbRoutes := []*pb.DNSRoute{}
		for _, routeCode := range domainRouteCodes {
			for _, r := range routes {
				if r.Code == routeCode {
					pbRoutes = append(pbRoutes, &pb.DNSRoute{
						Name: r.Name,
						Code: r.Code,
					})
					break
				}
			}
		}

		for _, ipAddress := range ipAddresses {
			ip := ipAddress.Ip
			if len(ip) == 0 {
				continue
			}
			if net.ParseIP(ip) == nil {
				continue
			}
			result = append(result, &pb.NodeDNSInfo{
				Id:     int64(node.Id),
				Name:   node.Name,
				IpAddr: ip,
				Routes: pbRoutes,
			})
		}
	}
	return &pb.FindAllEnabledNodesDNSWithNodeClusterIdResponse{Nodes: result}, nil
}

// FindEnabledNodeDNS 查找单个节点的域名解析信息
func (this *NodeService) FindEnabledNodeDNS(ctx context.Context, req *pb.FindEnabledNodeDNSRequest) (*pb.FindEnabledNodeDNSResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedNodeDAO.FindEnabledNodeDNS(tx, req.NodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledNodeDNSResponse{Node: nil}, nil
	}

	ipAddr, err := models.SharedNodeIPAddressDAO.FindFirstNodeAccessIPAddress(tx, int64(node.Id))
	if err != nil {
		return nil, err
	}

	clusterId := int64(node.ClusterId)
	clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId)
	if err != nil {
		return nil, err
	}
	if clusterDNS == nil {
		return &pb.FindEnabledNodeDNSResponse{Node: nil}, nil
	}

	dnsDomainId := int64(clusterDNS.DnsDomainId)
	dnsDomainName, err := dns.SharedDNSDomainDAO.FindDNSDomainName(tx, dnsDomainId)
	if err != nil {
		return nil, err
	}

	pbRoutes := []*pb.DNSRoute{}
	if dnsDomainId > 0 {
		routeCodes, err := node.DNSRouteCodesForDomainId(dnsDomainId)
		if err != nil {
			return nil, err
		}

		for _, routeCode := range routeCodes {
			routeName, err := dns.SharedDNSDomainDAO.FindDomainRouteName(tx, dnsDomainId, routeCode)
			if err != nil {
				return nil, err
			}
			pbRoutes = append(pbRoutes, &pb.DNSRoute{
				Name: routeName,
				Code: routeCode,
			})
		}
	}

	return &pb.FindEnabledNodeDNSResponse{
		Node: &pb.NodeDNSInfo{
			Id:                 int64(node.Id),
			Name:               node.Name,
			IpAddr:             ipAddr,
			Routes:             pbRoutes,
			NodeClusterId:      clusterId,
			NodeClusterDNSName: clusterDNS.DnsName,
			DnsDomainId:        dnsDomainId,
			DnsDomainName:      dnsDomainName,
		},
	}, nil
}

// UpdateNodeDNS 修改节点的DNS解析信息
func (this *NodeService) UpdateNodeDNS(ctx context.Context, req *pb.UpdateNodeDNSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedNodeDAO.FindEnabledNodeDNS(tx, req.NodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, errors.New("node not found")
	}

	routeCodeMap, err := node.DNSRouteCodes()
	if err != nil {
		return nil, err
	}
	if req.DnsDomainId > 0 && len(req.Routes) > 0 {
		routeCodeMap[req.DnsDomainId] = req.Routes
	}

	err = models.SharedNodeDAO.UpdateNodeDNS(tx, req.NodeId, routeCodeMap)
	if err != nil {
		return nil, err
	}

	// 修改IP
	if len(req.IpAddr) > 0 {
		ipAddrId, err := models.SharedNodeIPAddressDAO.FindFirstNodeAccessIPAddressId(tx, req.NodeId)
		if err != nil {
			return nil, err
		}
		if ipAddrId > 0 {
			err = models.SharedNodeIPAddressDAO.UpdateAddressIP(tx, ipAddrId, req.IpAddr)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = models.SharedNodeIPAddressDAO.CreateAddress(tx, req.NodeId, "DNS IP", req.IpAddr, true)
			if err != nil {
				return nil, err
			}
		}
	}

	return this.Success()
}

// CountAllEnabledNodesWithNodeRegionId 计算某个区域下的节点数量
func (this *NodeService) CountAllEnabledNodesWithNodeRegionId(ctx context.Context, req *pb.CountAllEnabledNodesWithNodeRegionIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeDAO.CountAllEnabledNodesWithRegionId(tx, req.NodeRegionId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindEnabledNodesWithIds 根据一组ID获取节点信息
func (this *NodeService) FindEnabledNodesWithIds(ctx context.Context, req *pb.FindEnabledNodesWithIdsRequest) (*pb.FindEnabledNodesWithIdsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedNodeDAO.FindEnabledNodesWithIds(tx, req.NodeIds)
	if err != nil {
		return nil, err
	}
	pbNodes := []*pb.Node{}
	for _, node := range nodes {
		connectedAPINodeIds, err := node.DecodeConnectedAPINodeIds()
		if err != nil {
			return nil, err
		}
		pbNodes = append(pbNodes, &pb.Node{
			Id:                  int64(node.Id),
			IsOn:                node.IsOn == 1,
			IsActive:            node.IsActive == 1,
			ConnectedAPINodeIds: connectedAPINodeIds,
		})
	}
	return &pb.FindEnabledNodesWithIdsResponse{Nodes: pbNodes}, nil
}

// CheckNodeLatestVersion 检查新版本
func (this *NodeService) CheckNodeLatestVersion(ctx context.Context, req *pb.CheckNodeLatestVersionRequest) (*pb.CheckNodeLatestVersionResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	deployFiles := installers.SharedDeployManager.LoadFiles()
	for _, file := range deployFiles {
		if file.OS == req.Os && file.Arch == req.Arch && stringutil.VersionCompare(file.Version, req.CurrentVersion) > 0 {
			return &pb.CheckNodeLatestVersionResponse{
				HasNewVersion: true,
				NewVersion:    file.Version,
			}, nil
		}
	}
	return &pb.CheckNodeLatestVersionResponse{HasNewVersion: false}, nil
}

// UpdateNodeUp 设置节点上线状态
func (this *NodeService) UpdateNodeUp(ctx context.Context, req *pb.UpdateNodeUpRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeDAO.UpdateNodeUp(tx, req.NodeId, req.IsUp)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
