package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns/dnsutils"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"strconv"
)

type NodeClusterService struct {
	BaseService
}

// CreateNodeCluster 创建集群
func (this *NodeClusterService) CreateNodeCluster(ctx context.Context, req *pb.CreateNodeClusterRequest) (*pb.CreateNodeClusterResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// 系统服务
	systemServices := map[string]maps.Map{}
	if len(req.SystemServicesJSON) > 0 {
		err = json.Unmarshal(req.SystemServicesJSON, &systemServices)
		if err != nil {
			return nil, err
		}
	}

	var clusterId int64
	err = this.RunTx(func(tx *dbs.Tx) error {
		// 缓存策略
		if req.HttpCachePolicyId <= 0 {
			policyId, err := models.SharedHTTPCachePolicyDAO.CreateDefaultCachePolicy(tx, req.Name)
			if err != nil {
				return err
			}
			req.HttpCachePolicyId = policyId
		}

		// WAF策略
		if req.HttpFirewallPolicyId <= 0 {
			policyId, err := models.SharedHTTPFirewallPolicyDAO.CreateDefaultFirewallPolicy(tx, req.Name)
			if err != nil {
				return err
			}
			req.HttpFirewallPolicyId = policyId
		}

		clusterId, err = models.SharedNodeClusterDAO.CreateCluster(tx, adminId, req.Name, req.NodeGrantId, req.InstallDir, req.DnsDomainId, req.DnsName, req.HttpCachePolicyId, req.HttpFirewallPolicyId, systemServices)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &pb.CreateNodeClusterResponse{NodeClusterId: clusterId}, nil
}

// UpdateNodeCluster 修改集群
func (this *NodeClusterService) UpdateNodeCluster(ctx context.Context, req *pb.UpdateNodeClusterRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateCluster(tx, req.NodeClusterId, req.Name, req.NodeGrantId, req.InstallDir, req.TimeZone, req.NodeMaxThreads, req.NodeTCPMaxConnections, req.AutoOpenPorts)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteNodeCluster 禁用集群
func (this *NodeClusterService) DeleteNodeCluster(ctx context.Context, req *pb.DeleteNodeClusterRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if req.NodeClusterId <= 0 {
		return this.Success()
	}

	tx := this.NullTx()

	// 转移节点
	err = models.SharedNodeDAO.TransferPrimaryClusterNodes(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	// 删除集群
	err = models.SharedNodeClusterDAO.DisableNodeCluster(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	// 删除相关任务
	err = models.SharedNodeTaskDAO.DeleteAllClusterTasks(tx, nodeconfigs.NodeRoleNode, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledNodeCluster 查找单个集群
func (this *NodeClusterService) FindEnabledNodeCluster(ctx context.Context, req *pb.FindEnabledNodeClusterRequest) (*pb.FindEnabledNodeClusterResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查用户是否有权限
	}

	tx := this.NullTx()

	cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	if cluster == nil {
		return &pb.FindEnabledNodeClusterResponse{}, nil
	}

	return &pb.FindEnabledNodeClusterResponse{NodeCluster: &pb.NodeCluster{
		Id:                    int64(cluster.Id),
		Name:                  cluster.Name,
		CreatedAt:             int64(cluster.CreatedAt),
		InstallDir:            cluster.InstallDir,
		NodeGrantId:           int64(cluster.GrantId),
		UniqueId:              cluster.UniqueId,
		Secret:                cluster.Secret,
		HttpCachePolicyId:     int64(cluster.CachePolicyId),
		HttpFirewallPolicyId:  int64(cluster.HttpFirewallPolicyId),
		DnsName:               cluster.DnsName,
		DnsDomainId:           int64(cluster.DnsDomainId),
		IsOn:                  cluster.IsOn,
		TimeZone:              cluster.TimeZone,
		NodeMaxThreads:        int32(cluster.NodeMaxThreads),
		NodeTCPMaxConnections: int32(cluster.NodeTCPMaxConnections),
		AutoOpenPorts:         cluster.AutoOpenPorts == 1,
	}}, nil
}

// FindAPINodesWithNodeCluster 查找集群的API节点信息
func (this *NodeClusterService) FindAPINodesWithNodeCluster(ctx context.Context, req *pb.FindAPINodesWithNodeClusterRequest) (*pb.FindAPINodesWithNodeClusterResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.New("can not find cluster with id '" + strconv.FormatInt(req.NodeClusterId, 10) + "'")
	}

	result := &pb.FindAPINodesWithNodeClusterResponse{}
	result.UseAllAPINodes = cluster.UseAllAPINodes == 1

	apiNodeIds := []int64{}
	if models.IsNotNull(cluster.ApiNodes) {
		err = json.Unmarshal(cluster.ApiNodes, &apiNodeIds)
		if err != nil {
			return nil, err
		}
		if len(apiNodeIds) > 0 {
			apiNodes := []*pb.APINode{}
			for _, apiNodeId := range apiNodeIds {
				apiNode, err := models.SharedAPINodeDAO.FindEnabledAPINode(tx, apiNodeId, nil)
				if err != nil {
					return nil, err
				}
				apiNodeAddrs, err := apiNode.DecodeAccessAddrStrings()
				if err != nil {
					return nil, err
				}
				apiNodes = append(apiNodes, &pb.APINode{
					Id:            int64(apiNode.Id),
					IsOn:          apiNode.IsOn,
					NodeClusterId: int64(apiNode.ClusterId),
					Name:          apiNode.Name,
					Description:   apiNode.Description,
					AccessAddrs:   apiNodeAddrs,
				})
			}
			result.ApiNodes = apiNodes
		}
	}

	return result, nil
}

// FindAllEnabledNodeClusters 查找所有可用的集群
func (this *NodeClusterService) FindAllEnabledNodeClusters(ctx context.Context, req *pb.FindAllEnabledNodeClustersRequest) (*pb.FindAllEnabledNodeClustersResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.FindAllEnableClusters(tx)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:        int64(cluster.Id),
			Name:      cluster.Name,
			CreatedAt: int64(cluster.CreatedAt),
			UniqueId:  cluster.UniqueId,
			Secret:    cluster.Secret,
			IsOn:      cluster.IsOn,
		})
	}

	return &pb.FindAllEnabledNodeClustersResponse{
		NodeClusters: result,
	}, nil
}

// CountAllEnabledNodeClusters 计算所有集群数量
func (this *NodeClusterService) CountAllEnabledNodeClusters(ctx context.Context, req *pb.CountAllEnabledNodeClustersRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClusters(tx, req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledNodeClusters 列出单页集群
func (this *NodeClusterService) ListEnabledNodeClusters(ctx context.Context, req *pb.ListEnabledNodeClustersRequest) (*pb.ListEnabledNodeClustersResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.ListEnabledClusters(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:          int64(cluster.Id),
			Name:        cluster.Name,
			CreatedAt:   int64(cluster.CreatedAt),
			NodeGrantId: int64(cluster.GrantId),
			InstallDir:  cluster.InstallDir,
			UniqueId:    cluster.UniqueId,
			Secret:      cluster.Secret,
			DnsName:     cluster.DnsName,
			DnsDomainId: int64(cluster.DnsDomainId),
			IsOn:        cluster.IsOn,
			TimeZone:    cluster.TimeZone,
			IsPinned:    cluster.IsPinned,
		})
	}

	return &pb.ListEnabledNodeClustersResponse{NodeClusters: result}, nil
}

// FindNodeClusterHealthCheckConfig 查找集群的健康检查配置
func (this *NodeClusterService) FindNodeClusterHealthCheckConfig(ctx context.Context, req *pb.FindNodeClusterHealthCheckConfigRequest) (*pb.FindNodeClusterHealthCheckConfigResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	config, err := models.SharedNodeClusterDAO.FindClusterHealthCheckConfig(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindNodeClusterHealthCheckConfigResponse{HealthCheckJSON: configJSON}, nil
}

// UpdateNodeClusterHealthCheck 修改集群健康检查设置
func (this *NodeClusterService) UpdateNodeClusterHealthCheck(ctx context.Context, req *pb.UpdateNodeClusterHealthCheckRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateClusterHealthCheck(tx, req.NodeClusterId, req.HealthCheckJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ExecuteNodeClusterHealthCheck 执行健康检查
func (this *NodeClusterService) ExecuteNodeClusterHealthCheck(ctx context.Context, req *pb.ExecuteNodeClusterHealthCheckRequest) (*pb.ExecuteNodeClusterHealthCheckResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	executor := tasks.NewHealthCheckExecutor(req.NodeClusterId)
	results, err := executor.Run()
	if err != nil {
		return nil, err
	}
	pbResults := []*pb.ExecuteNodeClusterHealthCheckResponse_Result{}
	for _, result := range results {
		pbResults = append(pbResults, &pb.ExecuteNodeClusterHealthCheckResponse_Result{
			Node: &pb.Node{
				Id:   int64(result.Node.Id),
				Name: result.Node.Name,
			},
			NodeAddr: result.NodeAddr,
			IsOk:     result.IsOk,
			Error:    result.Error,
			CostMs:   types.Float32(result.CostMs),
		})
	}
	return &pb.ExecuteNodeClusterHealthCheckResponse{Results: pbResults}, nil
}

// CountAllEnabledNodeClustersWithNodeGrantId 计算使用某个认证的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithNodeGrantId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithNodeGrantIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClustersWithGrantId(tx, req.NodeGrantId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllEnabledNodeClustersWithNodeGrantId 查找使用某个认证的所有集群
func (this *NodeClusterService) FindAllEnabledNodeClustersWithNodeGrantId(ctx context.Context, req *pb.FindAllEnabledNodeClustersWithNodeGrantIdRequest) (*pb.FindAllEnabledNodeClustersWithNodeGrantIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersWithGrantId(tx, req.NodeGrantId)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:        int64(cluster.Id),
			Name:      cluster.Name,
			CreatedAt: int64(cluster.CreatedAt),
			UniqueId:  cluster.UniqueId,
			Secret:    cluster.Secret,
			IsOn:      cluster.IsOn,
		})
	}
	return &pb.FindAllEnabledNodeClustersWithNodeGrantIdResponse{NodeClusters: result}, nil
}

// FindEnabledNodeClusterDNS 查找集群的DNS配置
func (this *NodeClusterService) FindEnabledNodeClusterDNS(ctx context.Context, req *pb.FindEnabledNodeClusterDNSRequest) (*pb.FindEnabledNodeClusterDNSResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	dnsInfo, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, req.NodeClusterId, nil)
	if err != nil {
		return nil, err
	}
	if dnsInfo == nil {
		return &pb.FindEnabledNodeClusterDNSResponse{
			Name:     "",
			Domain:   nil,
			Provider: nil,
		}, nil
	}

	dnsConfig, err := dnsInfo.DecodeDNSConfig()
	if err != nil {
		return nil, err
	}

	if dnsInfo.DnsDomainId == 0 {
		return &pb.FindEnabledNodeClusterDNSResponse{
			Name:            dnsInfo.DnsName,
			Domain:          nil,
			Provider:        nil,
			NodesAutoSync:   dnsConfig.NodesAutoSync,
			ServersAutoSync: dnsConfig.ServersAutoSync,
			CnameRecords:    dnsConfig.CNameRecords,
			Ttl:             dnsConfig.TTL,
		}, nil
	}

	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, int64(dnsInfo.DnsDomainId), nil)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindEnabledNodeClusterDNSResponse{
			Name:     dnsInfo.DnsName,
			Domain:   nil,
			Provider: nil,
		}, nil
	}
	pbDomain := &pb.DNSDomain{
		Id:   int64(domain.Id),
		Name: domain.Name,
		IsOn: domain.IsOn,
	}

	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(domain.ProviderId))
	if err != nil {
		return nil, err
	}

	var defaultRoute = ""

	var pbProvider *pb.DNSProvider = nil
	if provider != nil {
		pbProvider = &pb.DNSProvider{
			Id:       int64(provider.Id),
			Name:     provider.Name,
			Type:     provider.Type,
			TypeName: dnsclients.FindProviderTypeName(provider.Type),
		}

		manager := dnsclients.FindProvider(provider.Type)
		if manager != nil {
			apiParams, err := provider.DecodeAPIParams()
			if err != nil {
				return nil, err
			}
			err = manager.Auth(apiParams)
			if err != nil {
				return nil, err
			}
			defaultRoute = manager.DefaultRoute()
		}
	}

	return &pb.FindEnabledNodeClusterDNSResponse{
		Name:            dnsInfo.DnsName,
		Domain:          pbDomain,
		Provider:        pbProvider,
		NodesAutoSync:   dnsConfig.NodesAutoSync,
		ServersAutoSync: dnsConfig.ServersAutoSync,
		CnameRecords:    dnsConfig.CNameRecords,
		Ttl:             dnsConfig.TTL,
		DefaultRoute:    defaultRoute,
	}, nil
}

// CountAllEnabledNodeClustersWithDNSProviderId 计算使用某个DNS服务商的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithDNSProviderId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithDNSProviderIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClustersWithDNSProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountAllEnabledNodeClustersWithDNSDomainId 计算使用某个DNS域名的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithDNSDomainId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithDNSDomainIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClustersWithDNSDomainId(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllEnabledNodeClustersWithDNSDomainId 查找使用某个域名的所有集群
func (this *NodeClusterService) FindAllEnabledNodeClustersWithDNSDomainId(ctx context.Context, req *pb.FindAllEnabledNodeClustersWithDNSDomainIdRequest) (*pb.FindAllEnabledNodeClustersWithDNSDomainIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersWithDNSDomainId(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:          int64(cluster.Id),
			Name:        cluster.Name,
			DnsName:     cluster.DnsName,
			DnsDomainId: int64(cluster.DnsDomainId),
			IsOn:        cluster.IsOn,
		})
	}
	return &pb.FindAllEnabledNodeClustersWithDNSDomainIdResponse{NodeClusters: result}, nil
}

// CheckNodeClusterDNSName 检查集群域名是否已经被使用
func (this *NodeClusterService) CheckNodeClusterDNSName(ctx context.Context, req *pb.CheckNodeClusterDNSNameRequest) (*pb.CheckNodeClusterDNSNameResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	exists, err := models.SharedNodeClusterDAO.ExistClusterDNSName(tx, req.DnsName, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.CheckNodeClusterDNSNameResponse{IsUsed: exists}, nil
}

// UpdateNodeClusterDNS 修改集群的域名设置
func (this *NodeClusterService) UpdateNodeClusterDNS(ctx context.Context, req *pb.UpdateNodeClusterDNSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateClusterDNS(tx, req.NodeClusterId, req.DnsName, req.DnsDomainId, req.NodesAutoSync, req.ServersAutoSync, req.CnameRecords, req.Ttl)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CheckNodeClusterDNSChanges 检查集群的DNS是否有变化
func (this *NodeClusterService) CheckNodeClusterDNSChanges(ctx context.Context, req *pb.CheckNodeClusterDNSChangesRequest) (*pb.CheckNodeClusterDNSChangesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	cluster, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, req.NodeClusterId, nil)
	if err != nil {
		return nil, err
	}

	if cluster == nil || len(cluster.DnsName) == 0 || cluster.DnsDomainId <= 0 {
		return &pb.CheckNodeClusterDNSChangesResponse{IsChanged: false}, nil
	}

	domainId := int64(cluster.DnsDomainId)
	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId, nil)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.CheckNodeClusterDNSChangesResponse{IsChanged: false}, nil
	}
	records, err := domain.DecodeRecords()
	if err != nil {
		return nil, err
	}

	defaultRoute, err := dnsutils.FindDefaultDomainRoute(tx, domain)
	if err != nil {
		return nil, err
	}

	service := &DNSDomainService{}
	changes, _, _, _, _, _, _, err := service.findClusterDNSChanges(cluster, records, domain.Name, defaultRoute)
	if err != nil {
		return nil, err
	}

	return &pb.CheckNodeClusterDNSChangesResponse{IsChanged: len(changes) > 0}, nil
}

// FindEnabledNodeClusterTOA 查找集群的TOA配置
func (this *NodeClusterService) FindEnabledNodeClusterTOA(ctx context.Context, req *pb.FindEnabledNodeClusterTOARequest) (*pb.FindEnabledNodeClusterTOAResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	tx := this.NullTx()

	config, err := models.SharedNodeClusterDAO.FindClusterTOAConfig(tx, req.NodeClusterId, nil)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledNodeClusterTOAResponse{ToaJSON: configJSON}, nil
}

// UpdateNodeClusterTOA 修改集群的TOA设置
func (this *NodeClusterService) UpdateNodeClusterTOA(ctx context.Context, req *pb.UpdateNodeClusterTOARequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateClusterTOA(tx, req.NodeClusterId, req.ToaJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountAllEnabledNodeClustersWithHTTPCachePolicyId 计算使用某个缓存策略的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithHTTPCachePolicyId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithHTTPCachePolicyIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledNodeClustersWithHTTPCachePolicyId(tx, req.HttpCachePolicyId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllEnabledNodeClustersWithHTTPCachePolicyId 查找使用缓存策略的所有集群
func (this *NodeClusterService) FindAllEnabledNodeClustersWithHTTPCachePolicyId(ctx context.Context, req *pb.FindAllEnabledNodeClustersWithHTTPCachePolicyIdRequest) (*pb.FindAllEnabledNodeClustersWithHTTPCachePolicyIdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.NodeCluster{}
	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledNodeClustersWithHTTPCachePolicyId(tx, req.HttpCachePolicyId)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:   int64(cluster.Id),
			Name: cluster.Name,
			IsOn: cluster.IsOn,
		})
	}
	return &pb.FindAllEnabledNodeClustersWithHTTPCachePolicyIdResponse{
		NodeClusters: result,
	}, nil
}

// CountAllEnabledNodeClustersWithHTTPFirewallPolicyId 计算使用某个WAF策略的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithHTTPFirewallPolicyId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithHTTPFirewallPolicyIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledNodeClustersWithHTTPFirewallPolicyId(tx, req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllEnabledNodeClustersWithHTTPFirewallPolicyId 查找使用WAF策略的所有集群
func (this *NodeClusterService) FindAllEnabledNodeClustersWithHTTPFirewallPolicyId(ctx context.Context, req *pb.FindAllEnabledNodeClustersWithHTTPFirewallPolicyIdRequest) (*pb.FindAllEnabledNodeClustersWithHTTPFirewallPolicyIdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.NodeCluster{}
	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledNodeClustersWithHTTPFirewallPolicyId(tx, req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:   int64(cluster.Id),
			Name: cluster.Name,
			IsOn: cluster.IsOn,
		})
	}
	return &pb.FindAllEnabledNodeClustersWithHTTPFirewallPolicyIdResponse{
		NodeClusters: result,
	}, nil
}

// UpdateNodeClusterHTTPCachePolicyId 修改集群的缓存策略
func (this *NodeClusterService) UpdateNodeClusterHTTPCachePolicyId(ctx context.Context, req *pb.UpdateNodeClusterHTTPCachePolicyIdRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateNodeClusterHTTPCachePolicyId(tx, req.NodeClusterId, req.HttpCachePolicyId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNodeClusterHTTPFirewallPolicyId 修改集群的WAF策略
func (this *NodeClusterService) UpdateNodeClusterHTTPFirewallPolicyId(ctx context.Context, req *pb.UpdateNodeClusterHTTPFirewallPolicyIdRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateNodeClusterHTTPFirewallPolicyId(tx, req.NodeClusterId, req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateNodeClusterSystemService 修改集群的系统服务设置
func (this *NodeClusterService) UpdateNodeClusterSystemService(ctx context.Context, req *pb.UpdateNodeClusterSystemServiceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	params := maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()
	err = models.SharedNodeClusterDAO.UpdateNodeClusterSystemService(tx, req.NodeClusterId, req.Type, params)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindNodeClusterSystemService 查找集群的系统服务设置
func (this *NodeClusterService) FindNodeClusterSystemService(ctx context.Context, req *pb.FindNodeClusterSystemServiceRequest) (*pb.FindNodeClusterSystemServiceResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	params, err := models.SharedNodeClusterDAO.FindNodeClusterSystemServiceParams(tx, req.NodeClusterId, req.Type)
	if err != nil {
		return nil, err
	}
	if params == nil {
		params = maps.Map{}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return &pb.FindNodeClusterSystemServiceResponse{ParamsJSON: paramsJSON}, nil
}

// FindFreePortInNodeCluster 获取集群中可以使用的端口
func (this *NodeClusterService) FindFreePortInNodeCluster(ctx context.Context, req *pb.FindFreePortInNodeClusterRequest) (*pb.FindFreePortInNodeClusterResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	globalConfig, err := models.SharedSysSettingDAO.ReadGlobalConfig(tx)
	if err != nil {
		return nil, err
	}

	// 检查端口
	portMin := globalConfig.TCPAll.PortRangeMin
	portMax := globalConfig.TCPAll.PortRangeMax
	denyPorts := globalConfig.TCPAll.DenyPorts

	if portMin == 0 && portMax == 0 {
		portMin = 10_000
		portMax = 40_000
	}
	if portMin < 1024 {
		portMin = 10_000
	}
	if portMin > 65534 {
		portMin = 65534
	}
	if portMax < 1024 {
		portMax = 30_000
	}
	if portMax > 65534 {
		portMax = 65534
	}

	if portMin > portMax {
		portMax, portMin = portMin, portMax
	}

	// 最多尝试N次
	for i := 0; i < 60; i++ {
		port := rands.Int(portMin, portMax)
		if len(denyPorts) > 0 && lists.ContainsInt(denyPorts, port) {
			continue
		}

		isUsing, err := models.SharedServerDAO.CheckPortIsUsing(tx, req.NodeClusterId, req.ProtocolFamily, port, 0, "")
		if err != nil {
			return nil, err
		}
		if !isUsing {
			return &pb.FindFreePortInNodeClusterResponse{Port: int32(port)}, nil
		}
	}

	return nil, errors.New("can not find random port")
}

// CheckPortIsUsingInNodeCluster 检查端口是否已经被使用
func (this *NodeClusterService) CheckPortIsUsingInNodeCluster(ctx context.Context, req *pb.CheckPortIsUsingInNodeClusterRequest) (*pb.CheckPortIsUsingInNodeClusterResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	isUsing, err := models.SharedServerDAO.CheckPortIsUsing(tx, req.NodeClusterId, req.ProtocolFamily, int(req.Port), req.ExcludeServerId, req.ExcludeProtocol)
	if err != nil {
		return nil, err
	}
	return &pb.CheckPortIsUsingInNodeClusterResponse{IsUsing: isUsing}, nil
}

// FindLatestNodeClusters 查找最近访问的集群
func (this *NodeClusterService) FindLatestNodeClusters(ctx context.Context, req *pb.FindLatestNodeClustersRequest) (*pb.FindLatestNodeClustersResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	clusters, err := models.SharedNodeClusterDAO.FindLatestNodeClusters(tx, req.Size)
	if err != nil {
		return nil, err
	}
	pbClusters := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		pbClusters = append(pbClusters, &pb.NodeCluster{
			Id:   int64(cluster.Id),
			Name: cluster.Name,
			IsOn: cluster.IsOn,
		})
	}
	return &pb.FindLatestNodeClustersResponse{NodeClusters: pbClusters}, nil
}

// FindEnabledNodeClusterConfigInfo 取得集群的配置概要信息
func (this *NodeClusterService) FindEnabledNodeClusterConfigInfo(ctx context.Context, req *pb.FindEnabledNodeClusterConfigInfoRequest) (*pb.FindEnabledNodeClusterConfigInfoResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return &pb.FindEnabledNodeClusterConfigInfoResponse{}, nil
	}

	var result = &pb.FindEnabledNodeClusterConfigInfoResponse{}

	// health check
	if models.IsNotNull(cluster.HealthCheck) {
		healthCheckConfig := &serverconfigs.HealthCheckConfig{}
		err = json.Unmarshal(cluster.HealthCheck, healthCheckConfig)
		if err != nil {
			return nil, err
		}
		result.HealthCheckIsOn = healthCheckConfig.IsOn
	}

	// firewall actions
	countFirewallActions, err := models.SharedNodeClusterFirewallActionDAO.CountAllEnabledFirewallActions(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result.HasFirewallActions = countFirewallActions > 0

	// thresholds
	countThresholds, err := models.SharedNodeThresholdDAO.CountAllEnabledThresholds(tx, "node", req.NodeClusterId, 0)
	if err != nil {
		return nil, err
	}
	result.HasThresholds = countThresholds > 0

	// message receivers
	countReceivers, err := models.SharedMessageReceiverDAO.CountAllEnabledReceivers(tx, nodeconfigs.NodeRoleNode, req.NodeClusterId, 0, 0, "")
	if err != nil {
		return nil, err
	}
	result.HasMessageReceivers = countReceivers > 0

	// toa
	if models.IsNotNull(cluster.Toa) {
		var toaConfig = &nodeconfigs.TOAConfig{}
		err = json.Unmarshal(cluster.Toa, toaConfig)
		if err != nil {
			return nil, err
		}
		result.IsTOAEnabled = toaConfig.IsOn
	}

	// metric items
	countMetricItems, err := models.SharedNodeClusterMetricItemDAO.CountAllClusterItems(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result.HasMetricItems = countMetricItems > 0

	// webp
	if models.IsNotNull(cluster.Webp) {
		var webpPolicy = &nodeconfigs.WebPImagePolicy{}
		err = json.Unmarshal(cluster.Webp, webpPolicy)
		if err != nil {
			return nil, err
		}
		result.WebpIsOn = webpPolicy.IsOn
	} else {
		result.WebpIsOn = nodeconfigs.DefaultWebPImagePolicy.IsOn
	}

	return result, nil
}

// UpdateNodeClusterPinned 设置集群是否置顶
func (this *NodeClusterService) UpdateNodeClusterPinned(ctx context.Context, req *pb.UpdateNodeClusterPinnedRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeClusterDAO.UpdateClusterIsPinned(tx, req.NodeClusterId, req.IsPinned)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledNodeClusterWebPPolicy 读取集群WebP策略
func (this *NodeClusterService) FindEnabledNodeClusterWebPPolicy(ctx context.Context, req *pb.FindEnabledNodeClusterWebPPolicyRequest) (*pb.FindEnabledNodeClusterWebPPolicyResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	webpPolicy, err := models.SharedNodeClusterDAO.FindClusterWebPPolicy(tx, req.NodeClusterId, nil)
	if err != nil {
		return nil, err
	}
	webpPolicyJSON, err := json.Marshal(webpPolicy)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledNodeClusterWebPPolicyResponse{
		WebpPolicyJSON: webpPolicyJSON,
	}, nil
}

// UpdateNodeClusterWebPPolicy 设置集群WebP策略
func (this *NodeClusterService) UpdateNodeClusterWebPPolicy(ctx context.Context, req *pb.UpdateNodeClusterWebPPolicyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var webpPolicy = &nodeconfigs.WebPImagePolicy{}
	err = json.Unmarshal(req.WebpPolicyJSON, webpPolicy)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeClusterDAO.UpdateClusterWebPPolicy(tx, req.NodeClusterId, webpPolicy)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
