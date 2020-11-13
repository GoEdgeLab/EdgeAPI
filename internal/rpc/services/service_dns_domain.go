package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// DNS域名相关服务
type DNSDomainService struct {
}

// 创建域名
func (this *DNSDomainService) CreateDNSDomain(ctx context.Context, req *pb.CreateDNSDomainRequest) (*pb.CreateDNSDomainResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	domainId, err := models.SharedDNSDomainDAO.CreateDomain(req.DnsProviderId, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateDNSDomainResponse{DnsDomainId: domainId}, nil
}

// 修改域名
func (this *DNSDomainService) UpdateDNSDomain(ctx context.Context, req *pb.UpdateDNSDomainRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedDNSDomainDAO.UpdateDomain(req.DnsDomainId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return rpcutils.Success()
}

// 删除域名
func (this *DNSDomainService) DeleteDNSDomain(ctx context.Context, req *pb.DeleteDNSDomainRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedDNSDomainDAO.DisableDNSDomain(req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	return rpcutils.Success()
}

// 查询单个域名完整信息
func (this *DNSDomainService) FindEnabledDNSDomain(ctx context.Context, req *pb.FindEnabledDNSDomainRequest) (*pb.FindEnabledDNSDomainResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	domain, err := models.SharedDNSDomainDAO.FindEnabledDNSDomain(req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindEnabledDNSDomainResponse{DnsDomain: nil}, nil
	}

	pbDomain, err := this.convertDomainToPB(domain)
	return &pb.FindEnabledDNSDomainResponse{DnsDomain: pbDomain}, nil
}

// 查询单个域名基础信息
func (this *DNSDomainService) FindEnabledBasicDNSDomain(ctx context.Context, req *pb.FindEnabledBasicDNSDomainRequest) (*pb.FindEnabledBasicDNSDomainResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	domain, err := models.SharedDNSDomainDAO.FindEnabledDNSDomain(req.DnsDomainId)
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

// 计算服务商下的域名数量
func (this *DNSDomainService) CountAllEnabledDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.CountAllEnabledDNSDomainsWithDNSProviderIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedDNSDomainDAO.CountAllEnabledDomainsWithProviderId(req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return &pb.RPCCountResponse{Count: count}, nil
}

// 列出服务商下的所有域名
func (this *DNSDomainService) FindAllEnabledDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.FindAllEnabledDNSDomainsWithDNSProviderIdRequest) (*pb.FindAllEnabledDNSDomainsWithDNSProviderIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	domains, err := models.SharedDNSDomainDAO.FindAllEnabledDomainsWithProviderId(req.DnsProviderId)
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

// 列出服务商下的所有域名基本信息
func (this *DNSDomainService) FindAllEnabledBasicDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.FindAllEnabledBasicDNSDomainsWithDNSProviderIdRequest) (*pb.FindAllEnabledBasicDNSDomainsWithDNSProviderIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	domains, err := models.SharedDNSDomainDAO.FindAllEnabledDomainsWithProviderId(req.DnsProviderId)
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

// 同步域名数据
func (this *DNSDomainService) SyncDNSDomainData(ctx context.Context, req *pb.SyncDNSDomainDataRequest) (*pb.SyncDNSDomainDataResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	// 检查集群设置完整性
	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersWithDNSDomainId(req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		if len(cluster.DnsName) == 0 {
			return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "有问题需要修复", ShouldFix: true}, nil
		}
		nodes, err := models.SharedNodeDAO.FindAllEnabledNodesWithClusterId(int64(cluster.Id))
		if err != nil {
			return nil, err
		}
		for _, node := range nodes {
			if node.IsOn == 0 {
				continue
			}
			ipAddress, err := models.SharedNodeIPAddressDAO.FindFirstNodeIPAddress(int64(node.Id))
			if err != nil {
				return nil, err
			}
			if len(ipAddress) == 0 {
				return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "有问题需要修复", ShouldFix: true}, nil
			}
			route, err := node.DNSRoute(req.DnsDomainId)
			if err != nil {
				return nil, err
			}
			if len(route) == 0 {
				return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "有问题需要修复", ShouldFix: true}, nil
			}
		}
	}

	// 检查服务设置完整性
	servers, err := models.SharedServerDAO.FindAllServersToFixWithDNSDomainId(req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	if len(servers) > 0 {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "有问题需要修复", ShouldFix: true}, nil
	}

	// 域名信息
	domain, err := models.SharedDNSDomainDAO.FindEnabledDNSDomain(req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "找不到要操作的域名"}, nil
	}
	domainId := int64(domain.Id)
	domainName := domain.Name

	// 服务商信息
	provider, err := models.SharedDNSProviderDAO.FindEnabledDNSProvider(int64(domain.ProviderId))
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

	// 线路
	routes, err := manager.GetRoutes(domainName)
	if err != nil {
		return &pb.SyncDNSDomainDataResponse{IsOk: false, Error: "获取线路失败：" + err.Error()}, nil
	}
	routesJSON, err := json.Marshal(routes)
	if err != nil {
		return nil, err
	}
	err = models.SharedDNSDomainDAO.UpdateDomainRoutes(domainId, routesJSON)
	if err != nil {
		return nil, err
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
	err = models.SharedDNSDomainDAO.UpdateDomainRecords(domainId, recordsJSON)
	if err != nil {
		return nil, err
	}

	// 修正集群域名
	// TODO

	// 修正服务域名
	// TODO

	return &pb.SyncDNSDomainDataResponse{
		IsOk: true,
	}, nil
}

// 查看支持的线路
func (this *DNSDomainService) FindAllDNSDomainRoutes(ctx context.Context, req *pb.FindAllDNSDomainRoutesRequest) (*pb.FindAllDNSDomainRoutesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	routes, err := models.SharedDNSDomainDAO.FindDomainRoutes(req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	return &pb.FindAllDNSDomainRoutesResponse{Routes: routes}, nil
}

// 转换域名信息
func (this *DNSDomainService) convertDomainToPB(domain *models.DNSDomain) (*pb.DNSDomain, error) {
	domainId := int64(domain.Id)

	records := []*dnsclients.Record{}
	if len(domain.Records) > 0 && domain.Records != "null" {
		err := json.Unmarshal([]byte(domain.Records), &records)
		if err != nil {
			return nil, err
		}
	}
	recordsMapping := map[string][]*dnsclients.Record{} // name_type => *Record
	for _, record := range records {
		key := record.Name + "_" + record.Type
		_, ok := recordsMapping[key]
		if ok {
			recordsMapping[key] = append(recordsMapping[key], record)
		} else {
			recordsMapping[key] = []*dnsclients.Record{record}
		}
	}

	// 集群域名
	clusterRecords := []*pb.DNSRecord{}
	allClusterResolved := true
	{
		// 检查是否所有的集群都已经被解析
		clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersWithDNSDomainId(domainId)
		if err != nil {
			return nil, err
		}
		for _, cluster := range clusters {
			clusterId := int64(cluster.Id)
			dnsName := cluster.DnsName

			// 子节点
			nodes, err := models.SharedNodeDAO.FindAllEnabledNodesWithClusterId(clusterId)
			if err != nil {
				return nil, err
			}
			nodeMapping := map[string]*models.Node{} // name_type_value_route => *Node
			for _, node := range nodes {
				if node.IsOn == 0 {
					continue
				}

				ipAddr, err := models.SharedNodeIPAddressDAO.FindFirstNodeIPAddress(int64(node.Id))
				if err != nil {
					return nil, err
				}
				route, err := node.DNSRoute(domainId)
				if err != nil {
					return nil, err
				}
				nodeMapping[dnsName+"_A_"+ipAddr+"_"+route] = node
			}

			// 已有的记录
			nodeRecordsMapping := map[string]*dnsclients.Record{} // name_type_value_route => *Record
			nodeRecords, _ := recordsMapping[dnsName+"_A"]
			for _, record := range nodeRecords {
				key := record.Name + "_" + record.Type + "_" + record.Value + "_" + record.Route
				nodeRecordsMapping[key] = record
			}

			// 检查有无多余的子节点
			for key, record := range nodeRecordsMapping {
				_, ok := nodeMapping[key]
				if !ok {
					allClusterResolved = false
					continue
				}
				clusterRecords = append(clusterRecords, this.convertRecordToPB(record))
			}

			// 检查有无少的子节点
			for key := range nodeMapping {
				_, ok := nodeRecordsMapping[key]
				if !ok {
					allClusterResolved = false
					break
				}
			}
		}
	}

	// 服务域名
	serverRecords := []*pb.DNSRecord{}
	allServersResolved := true

	// 检查是否所有的服务都已经被解析
	{
		dnsNames, err := models.SharedServerDAO.FindAllServerDNSNamesWithDNSDomainId(domainId)
		if err != nil {
			return nil, err
		}
		for _, dnsName := range dnsNames {
			if len(dnsName) == 0 {
				allServersResolved = true
				continue
			}
			key := dnsName + "_CNAME"
			recordList, ok := recordsMapping[key]
			if !ok {
				allServersResolved = false
				continue
			}
			for _, record := range recordList {
				serverRecords = append(serverRecords, this.convertRecordToPB(record))
			}
		}
	}

	// 线路
	routes := []string{}
	if len(domain.Routes) > 0 && domain.Routes != "null" {
		err := json.Unmarshal([]byte(domain.Routes), &routes)
		if err != nil {
			return nil, err
		}
	}

	return &pb.DNSDomain{
		Id:                  int64(domain.Id),
		ProviderId:          int64(domain.ProviderId),
		Name:                domain.Name,
		IsOn:                domain.IsOn == 1,
		DataUpdatedAt:       int64(domain.DataUpdatedAt),
		ClusterRecords:      clusterRecords,
		AllClustersResolved: allClusterResolved,
		ServerRecords:       serverRecords,
		AllServersResolved:  allServersResolved,
		Routes:              routes,
	}, nil
}

// 转换域名记录信息
func (this *DNSDomainService) convertRecordToPB(record *dnsclients.Record) *pb.DNSRecord {
	return &pb.DNSRecord{
		Id:    record.Id,
		Name:  record.Name,
		Value: record.Value,
		Type:  record.Type,
		Route: record.Route,
	}
}
