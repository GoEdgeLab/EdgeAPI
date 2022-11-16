package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns/dnsutils"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"net"
)

// DNSDomainService DNS域名相关服务
type DNSDomainService struct {
	BaseService
}

// CreateDNSDomain 创建域名
func (this *DNSDomainService) CreateDNSDomain(ctx context.Context, req *pb.CreateDNSDomainRequest) (*pb.CreateDNSDomainResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 查询Provider
	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, errors.New("can not find provider")
	}
	apiParams, err := provider.DecodeAPIParams()
	if err != nil {
		return nil, err
	}

	domainId, err := dns.SharedDNSDomainDAO.CreateDomain(tx, adminId, userId, req.DnsProviderId, req.Name)
	if err != nil {
		return nil, err
	}

	// 更新数据，且不提示错误
	goman.New(func() {
		domainName := req.Name

		providerInterface := dnsclients.FindProvider(provider.Type, int64(provider.Id))
		if providerInterface == nil {
			return
		}
		err = providerInterface.Auth(apiParams)
		if err != nil {
			// 这里我们刻意不提示错误
			return
		}
		routes, err := providerInterface.GetRoutes(domainName)
		if err != nil {
			return
		}
		routesJSON, err := json.Marshal(routes)
		if err != nil {
			return
		}
		err = dns.SharedDNSDomainDAO.UpdateDomainRoutes(tx, domainId, routesJSON)
		if err != nil {
			return
		}

		records, err := providerInterface.GetRecords(domainName)
		if err != nil {
			return
		}
		recordsJSON, err := json.Marshal(records)
		if err != nil {
			return
		}
		err = dns.SharedDNSDomainDAO.UpdateDomainRecords(tx, domainId, recordsJSON)
		if err != nil {
			return
		}
	})

	return &pb.CreateDNSDomainResponse{DnsDomainId: domainId}, nil
}

// UpdateDNSDomain 修改域名
func (this *DNSDomainService) UpdateDNSDomain(ctx context.Context, req *pb.UpdateDNSDomainRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = dns.SharedDNSDomainDAO.UpdateDomain(tx, req.DnsDomainId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteDNSDomain 删除域名
func (this *DNSDomainService) DeleteDNSDomain(ctx context.Context, req *pb.DeleteDNSDomainRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = dns.SharedDNSDomainDAO.UpdateDomainIsDeleted(tx, req.DnsDomainId, true)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// RecoverDNSDomain 恢复删除的域名
func (this *DNSDomainService) RecoverDNSDomain(ctx context.Context, req *pb.RecoverDNSDomainRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = dns.SharedDNSDomainDAO.UpdateDomainIsDeleted(tx, req.DnsDomainId, false)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindDNSDomain 查询单个域名完整信息
func (this *DNSDomainService) FindDNSDomain(ctx context.Context, req *pb.FindDNSDomainRequest) (*pb.FindDNSDomainResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, req.DnsDomainId, nil)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindDNSDomainResponse{DnsDomain: nil}, nil
	}

	pbDomain, err := this.convertDomainToPB(tx, domain)
	return &pb.FindDNSDomainResponse{DnsDomain: pbDomain}, nil
}

// FindBasicDNSDomain 查询单个域名基础信息
func (this *DNSDomainService) FindBasicDNSDomain(ctx context.Context, req *pb.FindBasicDNSDomainRequest) (*pb.FindBasicDNSDomainResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, req.DnsDomainId, nil)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindBasicDNSDomainResponse{DnsDomain: nil}, nil
	}

	return &pb.FindBasicDNSDomainResponse{DnsDomain: &pb.DNSDomain{
		Id:         int64(domain.Id),
		Name:       domain.Name,
		IsOn:       domain.IsOn,
		ProviderId: int64(domain.ProviderId),
	}}, nil
}

// CountAllDNSDomainsWithDNSProviderId 计算服务商下的域名数量
func (this *DNSDomainService) CountAllDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.CountAllDNSDomainsWithDNSProviderIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := dns.SharedDNSDomainDAO.CountAllEnabledDomainsWithProviderId(tx, req.DnsProviderId, req.IsDeleted, !req.IsDown)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllDNSDomainsWithDNSProviderId 列出服务商下的所有域名
func (this *DNSDomainService) FindAllDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.FindAllDNSDomainsWithDNSProviderIdRequest) (*pb.FindAllDNSDomainsWithDNSProviderIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	domains, err := dns.SharedDNSDomainDAO.FindAllEnabledDomainsWithProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}

	result := []*pb.DNSDomain{}
	for _, domain := range domains {
		pbDomain, err := this.convertDomainToPB(tx, domain)
		if err != nil {
			return nil, err
		}
		result = append(result, pbDomain)
	}

	return &pb.FindAllDNSDomainsWithDNSProviderIdResponse{DnsDomains: result}, nil
}

// FindAllBasicDNSDomainsWithDNSProviderId 列出服务商下的所有域名基本信息
func (this *DNSDomainService) FindAllBasicDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.FindAllBasicDNSDomainsWithDNSProviderIdRequest) (*pb.FindAllBasicDNSDomainsWithDNSProviderIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	domains, err := dns.SharedDNSDomainDAO.FindAllEnabledDomainsWithProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}

	var result = []*pb.DNSDomain{}
	for _, domain := range domains {
		result = append(result, &pb.DNSDomain{
			Id:        int64(domain.Id),
			Name:      domain.Name,
			IsOn:      domain.IsOn,
			IsUp:      domain.IsUp,
			IsDeleted: domain.IsDeleted,
		})
	}

	return &pb.FindAllBasicDNSDomainsWithDNSProviderIdResponse{DnsDomains: result}, nil
}

// ListBasicDNSDomainsWithDNSProviderId 列出服务商下的单页域名信息
func (this *DNSDomainService) ListBasicDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.ListBasicDNSDomainsWithDNSProviderIdRequest) (*pb.ListDNSDomainsWithDNSProviderIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	domains, err := dns.SharedDNSDomainDAO.ListDomains(tx, req.DnsProviderId, req.IsDeleted, !req.IsDown, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var result = []*pb.DNSDomain{}
	for _, domain := range domains {
		pbDomain, err := this.convertDomainToPB(tx, domain)
		if err != nil {
			return nil, err
		}
		result = append(result, pbDomain)
	}

	return &pb.ListDNSDomainsWithDNSProviderIdResponse{DnsDomains: result}, nil
}

// SyncDNSDomainData 同步域名数据
func (this *DNSDomainService) SyncDNSDomainData(ctx context.Context, req *pb.SyncDNSDomainDataRequest) (*pb.SyncDNSDomainDataResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return this.syncClusterDNS(req)
}

// FindAllDNSDomainRoutes 查看支持的线路
func (this *DNSDomainService) FindAllDNSDomainRoutes(ctx context.Context, req *pb.FindAllDNSDomainRoutesRequest) (*pb.FindAllDNSDomainRoutesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	routes, err := dns.SharedDNSDomainDAO.FindDomainRoutes(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}

	pbRoutes := []*pb.DNSRoute{}
	for _, route := range routes {
		pbRoutes = append(pbRoutes, &pb.DNSRoute{
			Name: route.Name,
			Code: route.Code,
		})
	}

	return &pb.FindAllDNSDomainRoutesResponse{Routes: pbRoutes}, nil
}

// ExistAvailableDomains 判断是否有域名可选
func (this *DNSDomainService) ExistAvailableDomains(ctx context.Context, req *pb.ExistAvailableDomainsRequest) (*pb.ExistAvailableDomainsResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	exist, err := dns.SharedDNSDomainDAO.ExistAvailableDomains(tx)
	if err != nil {
		return nil, err
	}
	return &pb.ExistAvailableDomainsResponse{Exist: exist}, nil
}

// 转换域名信息
func (this *DNSDomainService) convertDomainToPB(tx *dbs.Tx, domain *dns.DNSDomain) (*pb.DNSDomain, error) {
	var domainId = int64(domain.Id)

	defaultRoute, err := dnsutils.FindDefaultDomainRoute(tx, domain)
	if err != nil {
		return nil, err
	}

	records := []*dnstypes.Record{}
	if models.IsNotNull(domain.Records) {
		err := json.Unmarshal(domain.Records, &records)
		if err != nil {
			return nil, err
		}
	}

	// 集群域名
	countNodeRecords := 0
	nodesChanged := false

	// 服务域名
	countServerRecords := 0
	serversChanged := false

	// 检查是否所有的集群都已经被解析
	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersWithDNSDomainId(tx, domainId)
	if err != nil {
		return nil, err
	}
	countClusters := len(clusters)
	countAllNodes1 := int64(0)
	countAllServers1 := int64(0)
	for _, cluster := range clusters {
		_, nodeRecords, serverRecords, countAllNodes, countAllServers, nodesChanged2, serversChanged2, err := this.findClusterDNSChanges(cluster, records, domain.Name, defaultRoute)
		if err != nil {
			return nil, err
		}
		countNodeRecords += len(nodeRecords)
		countServerRecords += len(serverRecords)
		countAllNodes1 += countAllNodes
		countAllServers1 += countAllServers
		if nodesChanged2 {
			nodesChanged = true
		}
		if serversChanged2 {
			serversChanged = true
		}
	}

	// 线路
	routes, err := domain.DecodeRoutes()
	if err != nil {
		return nil, err
	}
	pbRoutes := []*pb.DNSRoute{}
	for _, route := range routes {
		pbRoutes = append(pbRoutes, &pb.DNSRoute{
			Name: route.Name,
			Code: route.Code,
		})
	}

	return &pb.DNSDomain{
		Id:                 int64(domain.Id),
		ProviderId:         int64(domain.ProviderId),
		Name:               domain.Name,
		IsOn:               domain.IsOn,
		IsUp:               domain.IsUp,
		IsDeleted:          domain.IsDeleted,
		DataUpdatedAt:      int64(domain.DataUpdatedAt),
		CountNodeRecords:   int64(countNodeRecords),
		NodesChanged:       nodesChanged,
		CountServerRecords: int64(countServerRecords),
		ServersChanged:     serversChanged,
		Routes:             pbRoutes,
		CountNodeClusters:  int64(countClusters),
		CountAllNodes:      countAllNodes1,
		CountAllServers:    countAllServers1,
	}, nil
}

// 检查集群节点变化
func (this *DNSDomainService) findClusterDNSChanges(cluster *models.NodeCluster, records []*dnstypes.Record, domainName string, defaultRoute string) (result []maps.Map, doneNodeRecords []*dnstypes.Record, doneServerRecords []*dnstypes.Record, countAllNodes int64, countAllServers int64, nodesChanged bool, serversChanged bool, err error) {
	var clusterId = int64(cluster.Id)
	var clusterDnsName = cluster.DnsName
	var clusterDomain = clusterDnsName + "." + domainName
	dnsConfig, err := cluster.DecodeDNSConfig()
	if err != nil {
		return nil, nil, nil, 0, 0, false, false, err
	}
	if dnsConfig == nil {
		dnsConfig = dnsconfigs.DefaultClusterDNSConfig()
	}

	var tx = this.NullTx()

	// 自动设置的cname记录
	var ttl int32
	var cnameRecords = dnsConfig.CNAMERecords
	if dnsConfig.TTL > 0 {
		ttl = dnsConfig.TTL
	}

	// 节点域名
	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(tx, clusterId, true, dnsConfig != nil && dnsConfig.IncludingLnNodes)
	if err != nil {
		return nil, nil, nil, 0, 0, false, false, err
	}
	countAllNodes = int64(len(nodes))
	var nodeRecords = []*dnstypes.Record{}                // 之所以用数组再存一遍，是因为dnsName可能会重复
	var nodeRecordMapping = map[string]*dnstypes.Record{} // value_route => *Record
	for _, record := range records {
		if (record.Type == dnstypes.RecordTypeA || record.Type == dnstypes.RecordTypeAAAA) && record.Name == clusterDnsName {
			nodeRecords = append(nodeRecords, record)
			nodeRecordMapping[record.Value+"_"+record.Route] = record
		}
	}

	// 新增的节点域名
	var nodeKeys = []string{}
	var addingNodeRecordKeysMap = map[string]bool{} // clusterDnsName_type_ip_route
	for _, node := range nodes {
		ipAddresses, err := models.SharedNodeIPAddressDAO.FindNodeAccessAndUpIPAddresses(tx, int64(node.Id), nodeconfigs.NodeRoleNode)
		if err != nil {
			return nil, nil, nil, 0, 0, false, false, err
		}
		if len(ipAddresses) == 0 {
			continue
		}
		routeCodes, err := node.DNSRouteCodesForDomainId(int64(cluster.DnsDomainId))
		if err != nil {
			return nil, nil, nil, 0, 0, false, false, err
		}
		if len(routeCodes) == 0 {
			// 默认线路
			if len(defaultRoute) > 0 {
				routeCodes = []string{defaultRoute}
			} else {
				continue
			}
		}
		for _, route := range routeCodes {
			for _, ipAddress := range ipAddresses {
				ip := ipAddress.DNSIP()
				if len(ip) == 0 {
					continue
				}
				if net.ParseIP(ip) == nil {
					continue
				}
				var key = ip + "_" + route
				nodeKeys = append(nodeKeys, key)
				record, ok := nodeRecordMapping[key]
				if !ok {
					recordType := dnstypes.RecordTypeA
					if utils.IsIPv6(ip) {
						recordType = dnstypes.RecordTypeAAAA
					}

					// 避免添加重复的记录
					var fullKey = clusterDnsName + "_" + recordType + "_" + ip + "_" + route
					if addingNodeRecordKeysMap[fullKey] {
						continue
					}
					addingNodeRecordKeysMap[fullKey] = true

					result = append(result, maps.Map{
						"action": "create",
						"record": &dnstypes.Record{
							Id:    "",
							Name:  clusterDnsName,
							Type:  recordType,
							Value: ip,
							Route: route,
							TTL:   ttl,
						},
					})
					nodesChanged = true
				} else {
					doneNodeRecords = append(doneNodeRecords, record)
				}
			}
		}
	}

	// 多余的节点域名
	for _, record := range nodeRecords {
		key := record.Value + "_" + record.Route
		if !lists.ContainsString(nodeKeys, key) {
			nodesChanged = true
			result = append(result, maps.Map{
				"action": "delete",
				"record": record,
			})
		}
	}

	// 服务域名
	servers, err := models.SharedServerDAO.FindAllServersDNSWithClusterId(tx, clusterId)
	if err != nil {
		return nil, nil, nil, 0, 0, false, false, err
	}
	countAllServers = int64(len(servers))
	var serverRecords = []*dnstypes.Record{}             // 之所以用数组再存一遍，是因为dnsName可能会重复
	var serverRecordsMap = map[string]*dnstypes.Record{} // dnsName => *Record
	for _, record := range records {
		if record.Type == dnstypes.RecordTypeCNAME && record.Value == clusterDomain+"." {
			serverRecords = append(serverRecords, record)
			serverRecordsMap[record.Name] = record
		}
	}

	// 新增的域名
	var serverDNSNames = []string{}
	for _, server := range servers {
		var dnsName = server.DnsName
		if len(dnsName) == 0 {
			return nil, nil, nil, 0, 0, false, false, errors.New("server '" + numberutils.FormatInt64(int64(server.Id)) + "' 'dnsName' should not empty")
		}
		serverDNSNames = append(serverDNSNames, dnsName)
		record, ok := serverRecordsMap[dnsName]
		if !ok {
			serversChanged = true
			result = append(result, maps.Map{
				"action": "create",
				"record": &dnstypes.Record{
					Id:    "",
					Name:  dnsName,
					Type:  dnstypes.RecordTypeCNAME,
					Value: clusterDomain + ".",
					Route: "", // 注意这里为空，需要在执行过程中获取默认值
					TTL:   ttl,
				},
			})
		} else {
			doneServerRecords = append(doneServerRecords, record)
		}
	}

	// 自动设置的CNAME
	for _, cnameRecord := range cnameRecords {
		// 如果记录已存在，则跳过
		if lists.ContainsString(serverDNSNames, cnameRecord) {
			continue
		}

		serverDNSNames = append(serverDNSNames, cnameRecord)
		record, ok := serverRecordsMap[cnameRecord]
		if !ok {
			serversChanged = true
			result = append(result, maps.Map{
				"action": "create",
				"record": &dnstypes.Record{
					Id:    "",
					Name:  cnameRecord,
					Type:  dnstypes.RecordTypeCNAME,
					Value: clusterDomain + ".",
					Route: "", // 注意这里为空，需要在执行过程中获取默认值
					TTL:   ttl,
				},
			})
		} else {
			doneServerRecords = append(doneServerRecords, record)
		}
	}

	// 多余的域名
	for _, record := range serverRecords {
		if !lists.ContainsString(serverDNSNames, record.Name) {
			serversChanged = true
			result = append(result, maps.Map{
				"action": "delete",
				"record": record,
			})
		}
	}

	return
}

// 执行同步
func (this *DNSDomainService) syncClusterDNS(req *pb.SyncDNSDomainDataRequest) (*pb.SyncDNSDomainDataResponse, error) {
	var tx = this.NullTx()

	// 查询集群信息
	var err error
	var clusters = []*models.NodeCluster{}
	if req.NodeClusterId > 0 {
		cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, req.NodeClusterId)
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			return &pb.SyncDNSDomainDataResponse{
				IsOk:      false,
				Error:     "找不到要同步的集群",
				ShouldFix: false,
			}, nil
		}
		if int64(cluster.DnsDomainId) != req.DnsDomainId {
			return &pb.SyncDNSDomainDataResponse{
				IsOk:      false,
				Error:     "集群设置的域名和参数不符",
				ShouldFix: false,
			}, nil
		}
		clusters = append(clusters, cluster)
	} else {
		clusters, err = models.SharedNodeClusterDAO.FindAllEnabledClustersWithDNSDomainId(tx, req.DnsDomainId)
		if err != nil {
			return nil, err
		}
	}

	// 域名信息
	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, req.DnsDomainId, nil)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "找不到要操作的域名"}, nil
	}
	var domainId = int64(domain.Id)
	var domainName = domain.Name

	// 服务商信息
	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(domain.ProviderId))
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "域名没有设置服务商"}, nil
	}
	var apiParams = maps.Map{}
	if models.IsNotNull(provider.ApiParams) {
		err = json.Unmarshal(provider.ApiParams, &apiParams)
		if err != nil {
			return nil, err
		}
	}

	// 开始同步
	var manager = dnsclients.FindProvider(provider.Type, int64(provider.Id))
	if manager == nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "目前不支持'" + provider.Type + "'"}, nil
	}
	err = manager.Auth(apiParams)
	if err != nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "调用API认证失败：" + err.Error()}, nil
	}

	// 更新线路
	routes, err := manager.GetRoutes(domainName)
	if err != nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "获取线路失败：" + err.Error()}, nil
	}
	routesJSON, err := json.Marshal(routes)
	if err != nil {
		return nil, err
	}
	err = dns.SharedDNSDomainDAO.UpdateDomainRoutes(tx, domainId, routesJSON)
	if err != nil {
		return nil, err
	}

	// 检查集群设置
	for _, cluster := range clusters {
		issues, err := dnsutils.CheckClusterDNS(tx, cluster, req.CheckNodeIssues)
		if err != nil {
			return nil, err
		}
		if len(issues) > 0 {
			return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "发现问题需要修复", ShouldFix: true}, nil
		}
	}

	// 所有记录
	records, err := manager.GetRecords(domainName)
	if err != nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "获取域名解析记录失败：" + err.Error()}, nil
	}
	recordsJSON, err := json.Marshal(records)
	if err != nil {
		return nil, err
	}
	err = dns.SharedDNSDomainDAO.UpdateDomainRecords(tx, domainId, recordsJSON)
	if err != nil {
		return nil, err
	}

	// 对比变化
	var allChanges = []maps.Map{}
	for _, cluster := range clusters {
		changes, _, _, _, _, _, _, err := this.findClusterDNSChanges(cluster, records, domainName, manager.DefaultRoute())
		if err != nil {
			return nil, err
		}
		allChanges = append(allChanges, changes...)
	}
	for _, change := range allChanges {
		action := change.GetString("action")
		record := change.Get("record").(*dnstypes.Record)

		if len(record.Route) == 0 {
			record.Route = manager.DefaultRoute()
		}

		switch action {
		case "create":
			err = manager.AddRecord(domainName, record)
			if err != nil {
				return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "创建域名记录失败：" + err.Error()}, nil
			}
		case "delete":
			err = manager.DeleteRecord(domainName, record)
			if err != nil {
				return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "删除域名记录失败：" + err.Error()}, nil
			}
		}
	}

	// 重新更新记录
	if len(allChanges) > 0 {
		records, err := manager.GetRecords(domainName)
		if err != nil {
			return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "重新获取域名解析记录失败：" + err.Error()}, nil
		}
		recordsJSON, err := json.Marshal(records)
		if err != nil {
			return nil, err
		}
		err = dns.SharedDNSDomainDAO.UpdateDomainRecords(tx, domainId, recordsJSON)
		if err != nil {
			return nil, err
		}
	}

	return &pb.SyncDNSDomainDataResponse{
		IsOk: true,
	}, nil
}

// ExistDNSDomainRecord 检查域名是否在记录中
func (this *DNSDomainService) ExistDNSDomainRecord(ctx context.Context, req *pb.ExistDNSDomainRecordRequest) (*pb.ExistDNSDomainRecordResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	isOk, err := dns.SharedDNSDomainDAO.ExistDomainRecord(tx, req.DnsDomainId, req.Name, req.Type, req.Route, req.Value)
	if err != nil {
		return nil, err
	}
	return &pb.ExistDNSDomainRecordResponse{IsOk: isOk}, nil
}

// SyncDNSDomainsFromProvider 从服务商同步域名
func (this *DNSDomainService) SyncDNSDomainsFromProvider(ctx context.Context, req *pb.SyncDNSDomainsFromProviderRequest) (*pb.SyncDNSDomainsFromProviderResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, errors.New("can not find provider")
	}

	// 下线不存在的域名
	oldDomains, err := dns.SharedDNSDomainDAO.FindAllEnabledDomainsWithProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}

	dnsProvider := dnsclients.FindProvider(provider.Type, int64(provider.Id))
	if dnsProvider == nil {
		return nil, errors.New("provider type '" + provider.Type + "' is not supported yet")
	}

	params, err := provider.DecodeAPIParams()
	if err != nil {
		return nil, errors.New("decode params failed: " + err.Error())
	}
	err = dnsProvider.Auth(params)
	if err != nil {
		return nil, errors.New("auth failed: " + err.Error())
	}

	domainNames, err := dnsProvider.GetDomains()
	if err != nil {
		return nil, err
	}

	var hasChanges = false

	// 创建或上线域名
	for _, domainName := range domainNames {
		domain, err := dns.SharedDNSDomainDAO.FindEnabledDomainWithName(tx, req.DnsProviderId, domainName)
		if err != nil {
			return nil, err
		}
		if domain == nil {
			_, err = dns.SharedDNSDomainDAO.CreateDomain(tx, 0, 0, req.DnsProviderId, domainName)
			if err != nil {
				return nil, err
			}
			hasChanges = true
		} else if !domain.IsUp {
			err = dns.SharedDNSDomainDAO.UpdateDomainIsUp(tx, int64(domain.Id), true)
			if err != nil {
				return nil, err
			}
			hasChanges = true
		}
	}

	// 将老的域名置为下线
	for _, oldDomain := range oldDomains {
		var domainName = oldDomain.Name
		if oldDomain.IsUp && !lists.ContainsString(domainNames, domainName) {
			err = dns.SharedDNSDomainDAO.UpdateDomainIsUp(tx, int64(oldDomain.Id), false)
			if err != nil {
				return nil, err
			}
			hasChanges = true
		}
	}

	return &pb.SyncDNSDomainsFromProviderResponse{
		HasChanges: hasChanges,
	}, nil
}
