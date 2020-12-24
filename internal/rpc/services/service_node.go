package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/events"
	"github.com/TeaOSLab/EdgeAPI/internal/installers"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
)

func init() {
	dbs.OnReady(func() {
		go func() {
			service := &NodeService{}
			for nodeId := range events.NodeDNSChanges {
				err := service.notifyNodeDNSChanged(nodeId)
				if err != nil {
					logs.Println("[ERROR]change node dns: " + err.Error())
				}
			}
		}()
	})
}

// 边缘节点相关服务
type NodeService struct {
	BaseService
}

// 创建节点
func (this *NodeService) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	nodeId, err := models.SharedNodeDAO.CreateNode(adminId, req.Name, req.NodeClusterId, req.GroupId, req.RegionId)
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

	// 保存DNS相关
	if req.DnsDomainId > 0 && len(req.DnsRoutes) > 0 {
		err = models.SharedNodeDAO.UpdateNodeDNS(nodeId, map[int64][]string{
			req.DnsDomainId: req.DnsRoutes,
		})
		if err != nil {
			return nil, err
		}
	}

	// 同步DNS
	go func() {
		err := this.notifyNodeDNSChanged(nodeId)
		if err != nil {
			logs.Println("sync node DNS error: " + err.Error())
		}
	}()

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

	adminId, err := models.SharedNodeClusterDAO.FindClusterAdminId(clusterId)
	if err != nil {
		return nil, err
	}

	nodeId, err := models.SharedNodeDAO.CreateNode(adminId, req.Name, clusterId, 0, 0)
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
func (this *NodeService) CountAllEnabledNodes(ctx context.Context, req *pb.CountAllEnabledNodesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedNodeDAO.CountAllEnabledNodes()
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// 计算匹配的节点数量
func (this *NodeService) CountAllEnabledNodesMatch(ctx context.Context, req *pb.CountAllEnabledNodesMatchRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedNodeDAO.CountAllEnabledNodesMatch(req.NodeClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState), req.Keyword, req.GroupId, req.RegionId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 列出单页的节点
func (this *NodeService) ListEnabledNodesMatch(ctx context.Context, req *pb.ListEnabledNodesMatchRequest) (*pb.ListEnabledNodesMatchResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	dnsDomainId := int64(0)
	domainRoutes := []*dnsclients.Route{}
	if clusterDNS != nil {
		dnsDomainId = int64(clusterDNS.DnsDomainId)
		if clusterDNS.DnsDomainId > 0 {
			domainRoutes, err = models.SharedDNSDomainDAO.FindDomainRoutes(dnsDomainId)
			if err != nil {
				return nil, err
			}
		}
	}

	nodes, err := models.SharedNodeDAO.ListEnabledNodesMatch(req.Offset, req.Size, req.NodeClusterId, configutils.ToBoolState(req.InstallState), configutils.ToBoolState(req.ActiveState), req.Keyword, req.GroupId, req.RegionId)
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
				ErrorCode:  installStatus.ErrorCode,
				UpdatedAt:  installStatus.UpdatedAt,
			}
		}

		// 分组信息
		var pbGroup *pb.NodeGroup = nil
		if node.GroupId > 0 {
			group, err := models.SharedNodeGroupDAO.FindEnabledNodeGroup(int64(node.GroupId))
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
			region, err := models.SharedNodeRegionDAO.FindEnabledNodeRegion(int64(node.RegionId))
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

// 查找一个集群下的所有节点
func (this *NodeService) FindAllEnabledNodesWithClusterId(ctx context.Context, req *pb.FindAllEnabledNodesWithClusterIdRequest) (*pb.FindAllEnabledNodesWithClusterIdResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查权限
	}

	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesWithClusterId(req.NodeClusterId)
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

// 删除节点
func (this *NodeService) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.DisableNode(req.NodeId)
	if err != nil {
		return nil, err
	}

	// 同步DNS
	go func() {
		err := this.notifyNodeDNSChanged(req.NodeId)
		if err != nil {
			logs.Println("sync node DNS error: " + err.Error())
		}
	}()

	return this.Success()
}

// 修改节点
func (this *NodeService) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.UpdateNode(req.NodeId, req.Name, req.NodeClusterId, req.GroupId, req.RegionId, req.MaxCPU, req.IsOn)
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

	// 保存DNS相关
	if req.DnsDomainId > 0 && len(req.DnsRoutes) > 0 {
		err = models.SharedNodeDAO.UpdateNodeDNS(req.NodeId, map[int64][]string{
			req.DnsDomainId: req.DnsRoutes,
		})
		if err != nil {
			return nil, err
		}
	}

	// 同步DNS
	go func() {
		// TODO 只有状态变化的时候才需要同步
		err := this.notifyNodeDNSChanged(req.NodeId)
		if err != nil {
			logs.Println("sync node DNS error: " + err.Error())
		}
	}()

	return this.Success()
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
			ErrorCode:  installStatus.ErrorCode,
			UpdatedAt:  installStatus.UpdatedAt,
		}
	}

	// 分组信息
	var pbGroup *pb.NodeGroup = nil
	if node.GroupId > 0 {
		group, err := models.SharedNodeGroupDAO.FindEnabledNodeGroup(int64(node.GroupId))
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
		region, err := models.SharedNodeRegionDAO.FindEnabledNodeRegion(int64(node.RegionId))
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
		Login:         respLogin,
		InstallStatus: installStatusResult,
		MaxCPU:        types.Int32(node.MaxCPU),
		IsOn:          node.IsOn == 1,
		Group:         pbGroup,
		Region:        pbRegion,
	}}, nil
}

// 组合节点配置
func (this *NodeService) FindCurrentNodeConfig(ctx context.Context, req *pb.FindCurrentNodeConfigRequest) (*pb.FindCurrentNodeConfigResponse, error) {
	_ = req

	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	// 检查版本号
	currentVersion, err := models.SharedNodeDAO.FindNodeVersion(nodeId)
	if err != nil {
		return nil, err
	}
	if currentVersion == req.Version {
		return &pb.FindCurrentNodeConfigResponse{IsChanged: false}, nil
	}

	nodeConfig, err := models.SharedNodeDAO.ComposeNodeConfig(nodeId)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(nodeConfig)
	if err != nil {
		return nil, err
	}

	return &pb.FindCurrentNodeConfigResponse{IsChanged: true, NodeJSON: data}, nil
}

// 更新节点状态
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

	err = models.SharedNodeDAO.UpdateNodeStatus(nodeId, req.StatusJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 同步集群中的节点版本
func (this *NodeService) SyncNodesVersionWithCluster(ctx context.Context, req *pb.SyncNodesVersionWithClusterRequest) (*pb.SyncNodesVersionWithClusterResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.SyncNodeVersionsWithCluster(req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.SyncNodesVersionWithClusterResponse{}, nil
}

// 修改节点安装状态
func (this *NodeService) UpdateNodeIsInstalled(ctx context.Context, req *pb.UpdateNodeIsInstalledRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.UpdateNodeIsInstalled(req.NodeId, req.IsInstalled)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 安装节点
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

// 升级节点
func (this *NodeService) UpgradeNode(ctx context.Context, req *pb.UpgradeNodeRequest) (*pb.UpgradeNodeResponse, error) {
	// 校验节点
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.UpdateNodeIsInstalled(req.NodeId, false)
	if err != nil {
		return nil, err
	}

	// 检查状态
	installStatus, err := models.SharedNodeDAO.FindNodeInstallStatus(req.NodeId)
	if err != nil {
		return nil, err
	}
	if installStatus == nil {
		installStatus = &models.NodeInstallStatus{}
	}
	installStatus.IsOk = false
	installStatus.IsFinished = false
	err = models.SharedNodeDAO.UpdateNodeInstallStatus(req.NodeId, installStatus)
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

// 启动节点
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

	// 同步DNS
	go func() {
		err := this.notifyNodeDNSChanged(req.NodeId)
		if err != nil {
			logs.Println("sync node DNS error: " + err.Error())
		}
	}()

	return &pb.StartNodeResponse{IsOk: true}, nil
}

// 停止节点
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

	// 同步DNS
	go func() {
		err := this.notifyNodeDNSChanged(req.NodeId)
		if err != nil {
			logs.Println("sync node DNS error: " + err.Error())
		}
	}()

	return &pb.StopNodeResponse{IsOk: true}, nil
}

// 更改节点连接的API节点信息
func (this *NodeService) UpdateNodeConnectedAPINodes(ctx context.Context, req *pb.UpdateNodeConnectedAPINodesRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeDAO.UpdateNodeConnectedAPINodes(nodeId, req.ApiNodeIds)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return this.Success()
}

// 计算使用某个认证的节点数量
func (this *NodeService) CountAllEnabledNodesWithGrantId(ctx context.Context, req *pb.CountAllEnabledNodesWithGrantIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedNodeDAO.CountAllEnabledNodesWithGrantId(req.GrantId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
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
			StatusJSON:  []byte(node.Status),
			NodeCluster: &pb.NodeCluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
			IsOn: node.IsOn == 1,
		})
	}

	return &pb.FindAllEnabledNodesWithGrantIdResponse{Nodes: result}, nil
}

// 列出所有未安装的节点
func (this *NodeService) FindAllNotInstalledNodesWithClusterId(ctx context.Context, req *pb.FindAllNotInstalledNodesWithClusterIdRequest) (*pb.FindAllNotInstalledNodesWithClusterIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodes, err := models.SharedNodeDAO.FindAllNotInstalledNodesWithClusterId(req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result := []*pb.Node{}
	for _, node := range nodes {
		// 认证信息
		login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(int64(node.Id))
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
		addresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(int64(node.Id))
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
	return &pb.FindAllNotInstalledNodesWithClusterIdResponse{Nodes: result}, nil
}

// 计算需要升级的节点数量
func (this *NodeService) CountAllUpgradeNodesWithClusterId(ctx context.Context, req *pb.CountAllUpgradeNodesWithClusterIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	deployFiles := installers.SharedDeployManager.LoadFiles()
	total := int64(0)
	for _, deployFile := range deployFiles {
		count, err := models.SharedNodeDAO.CountAllLowerVersionNodesWithClusterId(req.NodeClusterId, deployFile.OS, deployFile.Arch, deployFile.Version)
		if err != nil {
			return nil, err
		}
		total += count
	}

	return this.SuccessCount(total)
}

// 列出所有需要升级的节点
func (this *NodeService) FindAllUpgradeNodesWithClusterId(ctx context.Context, req *pb.FindAllUpgradeNodesWithClusterIdRequest) (*pb.FindAllUpgradeNodesWithClusterIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// 获取当前能升级到的最新版本
	deployFiles := installers.SharedDeployManager.LoadFiles()
	result := []*pb.FindAllUpgradeNodesWithClusterIdResponse_NodeUpgrade{}
	for _, deployFile := range deployFiles {
		nodes, err := models.SharedNodeDAO.FindAllLowerVersionNodesWithClusterId(req.NodeClusterId, deployFile.OS, deployFile.Arch, deployFile.Version)
		if err != nil {
			return nil, err
		}
		for _, node := range nodes {
			// 认证信息
			login, err := models.SharedNodeLoginDAO.FindEnabledNodeLoginWithNodeId(int64(node.Id))
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
			addresses, err := models.SharedNodeIPAddressDAO.FindAllEnabledAddressesWithNode(int64(node.Id))
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

			result = append(result, &pb.FindAllUpgradeNodesWithClusterIdResponse_NodeUpgrade{
				Os:         deployFile.OS,
				Arch:       deployFile.Arch,
				OldVersion: status.BuildVersion,
				NewVersion: deployFile.Version,
				Node:       pbNode,
			})
		}
	}
	return &pb.FindAllUpgradeNodesWithClusterIdResponse{
		Nodes: result,
	}, nil
}

// 读取节点安装状态
func (this *NodeService) FindNodeInstallStatus(ctx context.Context, req *pb.FindNodeInstallStatusRequest) (*pb.FindNodeInstallStatusResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	installStatus, err := models.SharedNodeDAO.FindNodeInstallStatus(req.NodeId)
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

// 修改节点登录信息
func (this *NodeService) UpdateNodeLogin(ctx context.Context, req *pb.UpdateNodeLoginRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.Login.Id <= 0 {
		_, err := models.SharedNodeLoginDAO.CreateNodeLogin(req.NodeId, req.Login.Name, req.Login.Type, req.Login.Params)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedNodeLoginDAO.UpdateNodeLogin(req.Login.Id, req.Login.Name, req.Login.Type, req.Login.Params)

	return this.Success()
}

// 计算某个节点分组内的节点数量
func (this *NodeService) CountAllEnabledNodesWithNodeGroupId(ctx context.Context, req *pb.CountAllEnabledNodesWithNodeGroupIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedNodeDAO.CountAllEnabledNodesWithGroupId(req.NodeGroupId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 取得某个集群下的所有节点
func (this *NodeService) FindAllEnabledNodesDNSWithClusterId(ctx context.Context, req *pb.FindAllEnabledNodesDNSWithClusterIdRequest) (*pb.FindAllEnabledNodesDNSWithClusterIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	if clusterDNS == nil {
		return nil, errors.New("not found clusterId '" + numberutils.FormatInt64(req.NodeClusterId) + "'")
	}
	dnsDomainId := int64(clusterDNS.DnsDomainId)

	routes, err := models.SharedDNSDomainDAO.FindDomainRoutes(dnsDomainId)
	if err != nil {
		return nil, err
	}

	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeDNSInfo{}
	for _, node := range nodes {
		ipAddr, err := models.SharedNodeIPAddressDAO.FindFirstNodeIPAddress(int64(node.Id))
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

		result = append(result, &pb.NodeDNSInfo{
			Id:     int64(node.Id),
			Name:   node.Name,
			IpAddr: ipAddr,
			Routes: pbRoutes,
		})
	}
	return &pb.FindAllEnabledNodesDNSWithClusterIdResponse{Nodes: result}, nil
}

// 查找单个节点的域名解析信息
func (this *NodeService) FindEnabledNodeDNS(ctx context.Context, req *pb.FindEnabledNodeDNSRequest) (*pb.FindEnabledNodeDNSResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	node, err := models.SharedNodeDAO.FindEnabledNodeDNS(req.NodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledNodeDNSResponse{Node: nil}, nil
	}

	ipAddr, err := models.SharedNodeIPAddressDAO.FindFirstNodeIPAddress(int64(node.Id))
	if err != nil {
		return nil, err
	}

	clusterId := int64(node.ClusterId)
	clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(clusterId)
	if err != nil {
		return nil, err
	}
	if clusterDNS == nil {
		return &pb.FindEnabledNodeDNSResponse{Node: nil}, nil
	}

	dnsDomainId := int64(clusterDNS.DnsDomainId)
	dnsDomainName, err := models.SharedDNSDomainDAO.FindDNSDomainName(dnsDomainId)
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
			routeName, err := models.SharedDNSDomainDAO.FindDomainRouteName(dnsDomainId, routeCode)
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

// 修改节点的DNS解析信息
func (this *NodeService) UpdateNodeDNS(ctx context.Context, req *pb.UpdateNodeDNSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	node, err := models.SharedNodeDAO.FindEnabledNodeDNS(req.NodeId)
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

	err = models.SharedNodeDAO.UpdateNodeDNS(req.NodeId, routeCodeMap)
	if err != nil {
		return nil, err
	}

	// 修改IP
	if len(req.IpAddr) > 0 {
		ipAddrId, err := models.SharedNodeIPAddressDAO.FindFirstNodeIPAddressId(req.NodeId)
		if err != nil {
			return nil, err
		}
		if ipAddrId > 0 {
			err = models.SharedNodeIPAddressDAO.UpdateAddressIP(ipAddrId, req.IpAddr)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = models.SharedNodeIPAddressDAO.CreateAddress(req.NodeId, "DNS IP", req.IpAddr, true)
			if err != nil {
				return nil, err
			}
		}
	}

	return this.Success()
}

// 自动同步DNS状态
func (this *NodeService) notifyNodeDNSChanged(nodeId int64) error {
	clusterId, err := models.SharedNodeDAO.FindNodeClusterId(nodeId)
	if err != nil {
		return err
	}
	dnsInfo, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(clusterId)
	if err != nil {
		return err
	}
	if dnsInfo == nil {
		return nil
	}
	if len(dnsInfo.DnsName) == 0 || dnsInfo.DnsDomainId == 0 {
		return nil
	}
	dnsConfig, err := dnsInfo.DecodeDNSConfig()
	if err != nil {
		return err
	}
	if !dnsConfig.NodesAutoSync {
		return nil
	}

	// 执行同步
	domainService := &DNSDomainService{}
	resp, err := domainService.syncClusterDNS(&pb.SyncDNSDomainDataRequest{
		DnsDomainId:   int64(dnsInfo.DnsDomainId),
		NodeClusterId: clusterId,
	})
	if err != nil {
		return err
	}
	if !resp.IsOk {
		err = models.SharedMessageDAO.CreateClusterMessage(clusterId, models.MessageTypeClusterDNSSyncFailed, models.LevelError, "集群DNS同步失败："+resp.Error, nil)
		if err != nil {
			logs.Println("[NODE_SERVICE]" + err.Error())
		}
	}
	return nil
}

// 计算某个区域下的节点数量
func (this *NodeService) CountAllEnabledNodesWithNodeRegionId(ctx context.Context, req *pb.CountAllEnabledNodesWithNodeRegionIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedNodeDAO.CountAllEnabledNodesWithRegionId(req.NodeRegionId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}
