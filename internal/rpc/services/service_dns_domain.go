package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
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
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

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
	go func() {
		domainName := req.Name

		providerInterface := dnsclients.FindProvider(provider.Type)
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
	}()

	return &pb.CreateDNSDomainResponse{DnsDomainId: domainId}, nil
}

// UpdateDNSDomain 修改域名
func (this *DNSDomainService) UpdateDNSDomain(ctx context.Context, req *pb.UpdateDNSDomainRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = dns.SharedDNSDomainDAO.UpdateDomain(tx, req.DnsDomainId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteDNSDomain 删除域名
func (this *DNSDomainService) DeleteDNSDomain(ctx context.Context, req *pb.DeleteDNSDomainRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = dns.SharedDNSDomainDAO.DisableDNSDomain(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledDNSDomain 查询单个域名完整信息
func (this *DNSDomainService) FindEnabledDNSDomain(ctx context.Context, req *pb.FindEnabledDNSDomainRequest) (*pb.FindEnabledDNSDomainResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindEnabledDNSDomainResponse{DnsDomain: nil}, nil
	}

	pbDomain, err := this.convertDomainToPB(domain)
	return &pb.FindEnabledDNSDomainResponse{DnsDomain: pbDomain}, nil
}

// FindEnabledBasicDNSDomain 查询单个域名基础信息
func (this *DNSDomainService) FindEnabledBasicDNSDomain(ctx context.Context, req *pb.FindEnabledBasicDNSDomainRequest) (*pb.FindEnabledBasicDNSDomainResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindEnabledBasicDNSDomainResponse{DnsDomain: nil}, nil
	}

	return &pb.FindEnabledBasicDNSDomainResponse{DnsDomain: &pb.DNSDomain{
		Id:         int64(domain.Id),
		Name:       domain.Name,
		IsOn:       domain.IsOn == 1,
		ProviderId: int64(domain.ProviderId),
	}}, nil
}

// CountAllEnabledDNSDomainsWithDNSProviderId 计算服务商下的域名数量
func (this *DNSDomainService) CountAllEnabledDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.CountAllEnabledDNSDomainsWithDNSProviderIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := dns.SharedDNSDomainDAO.CountAllEnabledDomainsWithProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindAllEnabledDNSDomainsWithDNSProviderId 列出服务商下的所有域名
func (this *DNSDomainService) FindAllEnabledDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.FindAllEnabledDNSDomainsWithDNSProviderIdRequest) (*pb.FindAllEnabledDNSDomainsWithDNSProviderIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	domains, err := dns.SharedDNSDomainDAO.FindAllEnabledDomainsWithProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}

	result := []*pb.DNSDomain{}
	for _, domain := range domains {
		pbDomain, err := this.convertDomainToPB(domain)
		if err != nil {
			return nil, err
		}
		result = append(result, pbDomain)
	}

	return &pb.FindAllEnabledDNSDomainsWithDNSProviderIdResponse{DnsDomains: result}, nil
}

// FindAllEnabledBasicDNSDomainsWithDNSProviderId 列出服务商下的所有域名基本信息
func (this *DNSDomainService) FindAllEnabledBasicDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.FindAllEnabledBasicDNSDomainsWithDNSProviderIdRequest) (*pb.FindAllEnabledBasicDNSDomainsWithDNSProviderIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	domains, err := dns.SharedDNSDomainDAO.FindAllEnabledDomainsWithProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}

	result := []*pb.DNSDomain{}
	for _, domain := range domains {
		result = append(result, &pb.DNSDomain{
			Id:   int64(domain.Id),
			Name: domain.Name,
			IsOn: domain.IsOn == 1,
		})
	}

	return &pb.FindAllEnabledBasicDNSDomainsWithDNSProviderIdResponse{DnsDomains: result}, nil
}

// SyncDNSDomainData 同步域名数据
func (this *DNSDomainService) SyncDNSDomainData(ctx context.Context, req *pb.SyncDNSDomainDataRequest) (*pb.SyncDNSDomainDataResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	return this.syncClusterDNS(req)
}

// FindAllDNSDomainRoutes 查看支持的线路
func (this *DNSDomainService) FindAllDNSDomainRoutes(ctx context.Context, req *pb.FindAllDNSDomainRoutesRequest) (*pb.FindAllDNSDomainRoutesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

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
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	exist, err := dns.SharedDNSDomainDAO.ExistAvailableDomains(tx)
	if err != nil {
		return nil, err
	}
	return &pb.ExistAvailableDomainsResponse{Exist: exist}, nil
}

// 转换域名信息
func (this *DNSDomainService) convertDomainToPB(domain *dns.DNSDomain) (*pb.DNSDomain, error) {
	domainId := int64(domain.Id)

	records := []*dnstypes.Record{}
	if len(domain.Records) > 0 && domain.Records != "null" {
		err := json.Unmarshal([]byte(domain.Records), &records)
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

	tx := this.NullTx()

	// 检查是否所有的集群都已经被解析
	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersWithDNSDomainId(tx, domainId)
	if err != nil {
		return nil, err
	}
	countClusters := len(clusters)
	countAllNodes1 := int64(0)
	countAllServers1 := int64(0)
	for _, cluster := range clusters {
		_, nodeRecords, serverRecords, countAllNodes, countAllServers, nodesChanged2, serversChanged2, err := this.findClusterDNSChanges(cluster, records, domain.Name)
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
		IsOn:               domain.IsOn == 1,
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

// 转换域名记录信息
func (this *DNSDomainService) convertRecordToPB(record *dnstypes.Record) *pb.DNSRecord {
	return &pb.DNSRecord{
		Id:    record.Id,
		Name:  record.Name,
		Value: record.Value,
		Type:  record.Type,
		Route: record.Route,
	}
}

// 检查集群节点变化
func (this *DNSDomainService) findClusterDNSChanges(cluster *models.NodeCluster, records []*dnstypes.Record, domainName string) (result []maps.Map, doneNodeRecords []*dnstypes.Record, doneServerRecords []*dnstypes.Record, countAllNodes int64, countAllServers int64, nodesChanged bool, serversChanged bool, err error) {
	clusterId := int64(cluster.Id)
	clusterDnsName := cluster.DnsName
	clusterDomain := clusterDnsName + "." + domainName

	tx := this.NullTx()

	// 节点域名
	nodes, err := models.SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(tx, clusterId)
	if err != nil {
		return nil, nil, nil, 0, 0, false, false, err
	}
	countAllNodes = int64(len(nodes))
	nodeRecords := []*dnstypes.Record{}                // 之所以用数组再存一遍，是因为dnsName可能会重复
	nodeRecordMapping := map[string]*dnstypes.Record{} // value_route => *Record
	for _, record := range records {
		if (record.Type == dnstypes.RecordTypeA || record.Type == dnstypes.RecordTypeAAAA) && record.Name == clusterDnsName {
			nodeRecords = append(nodeRecords, record)
			nodeRecordMapping[record.Value+"_"+record.Route] = record
		}
	}

	// 新增的节点域名
	nodeKeys := []string{}
	for _, node := range nodes {
		ipAddresses, err := models.SharedNodeIPAddressDAO.FindNodeAccessIPAddresses(tx, int64(node.Id), nodeconfigs.NodeRoleNode)
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
			continue
		}
		for _, route := range routeCodes {
			for _, ipAddress := range ipAddresses {
				ip := ipAddress.Ip
				if len(ip) == 0 {
					continue
				}
				if net.ParseIP(ip) == nil {
					continue
				}
				key := ip + "_" + route
				nodeKeys = append(nodeKeys, key)
				record, ok := nodeRecordMapping[key]
				if !ok {
					recordType := dnstypes.RecordTypeA
					if utils.IsIPv6(ip) {
						recordType = dnstypes.RecordTypeAAAA
					}
					result = append(result, maps.Map{
						"action": "create",
						"record": &dnstypes.Record{
							Id:    "",
							Name:  clusterDnsName,
							Type:  recordType,
							Value: ip,
							Route: route,
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
	serverRecords := []*dnstypes.Record{}             // 之所以用数组再存一遍，是因为dnsName可能会重复
	serverRecordsMap := map[string]*dnstypes.Record{} // dnsName => *Record
	for _, record := range records {
		if record.Type == dnstypes.RecordTypeCNAME && record.Value == clusterDomain+"." {
			serverRecords = append(serverRecords, record)
			serverRecordsMap[record.Name] = record
		}
	}

	// 新增的域名
	serverDNSNames := []string{}
	for _, server := range servers {
		dnsName := server.DnsName
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
	tx := this.NullTx()

	// 查询集群信息
	var err error
	clusters := []*models.NodeCluster{}
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
	domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "找不到要操作的域名"}, nil
	}
	domainId := int64(domain.Id)
	domainName := domain.Name

	// 服务商信息
	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(domain.ProviderId))
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "域名没有设置服务商"}, nil
	}
	apiParams := maps.Map{}
	if len(provider.ApiParams) > 0 && provider.ApiParams != "null" {
		err = json.Unmarshal([]byte(provider.ApiParams), &apiParams)
		if err != nil {
			return nil, err
		}
	}

	// 开始同步
	manager := dnsclients.FindProvider(provider.Type)
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
		issues, err := models.SharedNodeClusterDAO.CheckClusterDNS(tx, cluster)
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
	allChanges := []maps.Map{}
	for _, cluster := range clusters {
		changes, _, _, _, _, _, _, err := this.findClusterDNSChanges(cluster, records, domainName)
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

		//logs.Println(action, record.Name, record.Type, record.Value, record.Route)
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

// 检查域名是否在记录中
func (this *DNSDomainService) ExistDNSDomainRecord(ctx context.Context, req *pb.ExistDNSDomainRecordRequest) (*pb.ExistDNSDomainRecordResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	isOk, err := dns.SharedDNSDomainDAO.ExistDomainRecord(tx, req.DnsDomainId, req.Name, req.Type, req.Route, req.Value)
	if err != nil {
		return nil, err
	}
	return &pb.ExistDNSDomainRecordResponse{IsOk: isOk}, nil
}
