package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"strconv"
)

const (
	NodeClusterStateEnabled  = 1 // 已启用
	NodeClusterStateDisabled = 0 // 已禁用
)

type NodeClusterDAO dbs.DAO

func NewNodeClusterDAO() *NodeClusterDAO {
	return dbs.NewDAO(&NodeClusterDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeClusters",
			Model:  new(NodeCluster),
			PkName: "id",
		},
	}).(*NodeClusterDAO)
}

var SharedNodeClusterDAO *NodeClusterDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeClusterDAO = NewNodeClusterDAO()
	})
}

// EnableNodeCluster 启用条目
func (this *NodeClusterDAO) EnableNodeCluster(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeClusterStateEnabled).
		Update()
	return err
}

// DisableNodeCluster 禁用条目
func (this *NodeClusterDAO) DisableNodeCluster(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeClusterStateDisabled).
		Update()
	return err
}

// FindEnabledNodeCluster 查找集群
func (this *NodeClusterDAO) FindEnabledNodeCluster(tx *dbs.Tx, id int64) (*NodeCluster, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeClusterStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeCluster), err
}

// FindEnabledClusterIdWithUniqueId 根据UniqueId获取ID
// TODO 增加缓存
func (this *NodeClusterDAO) FindEnabledClusterIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("uniqueId", uniqueId).
		ResultPk().
		FindInt64Col(0)
}

// FindNodeClusterName 根据主键查找名称
func (this *NodeClusterDAO) FindNodeClusterName(tx *dbs.Tx, clusterId int64) (string, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("name").
		FindStringCol("")
}

// FindAllEnableClusters 查找所有可用的集群
func (this *NodeClusterDAO) FindAllEnableClusters(tx *dbs.Tx) (result []*NodeCluster, err error) {
	_, err = this.Query(tx).
		State(NodeClusterStateEnabled).
		Slice(&result).
		Desc("order").
		DescPk().
		FindAll()
	return
}

// FindAllEnableClusterIds 查找所有可用的集群Ids
func (this *NodeClusterDAO) FindAllEnableClusterIds(tx *dbs.Tx) (result []int64, err error) {
	ones, err := this.Query(tx).
		State(NodeClusterStateEnabled).
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		result = append(result, int64(one.(*NodeCluster).Id))
	}
	return
}

// CreateCluster 创建集群
func (this *NodeClusterDAO) CreateCluster(tx *dbs.Tx, adminId int64, name string, grantId int64, installDir string, dnsDomainId int64, dnsName string, cachePolicyId int64, httpFirewallPolicyId int64, systemServices map[string]maps.Map) (clusterId int64, err error) {
	uniqueId, err := this.GenUniqueId(tx)
	if err != nil {
		return 0, err
	}

	secret := rands.String(32)
	err = SharedApiTokenDAO.CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleCluster)
	if err != nil {
		return 0, err
	}

	op := NewNodeClusterOperator()
	op.AdminId = adminId
	op.Name = name
	op.GrantId = grantId
	op.InstallDir = installDir

	// DNS设置
	op.DnsDomainId = dnsDomainId
	op.DnsName = dnsName
	dnsConfig := &dnsconfigs.ClusterDNSConfig{
		NodesAutoSync:   true,
		ServersAutoSync: true,
		CNameRecords:    []string{},
		TTL:             0,
	}
	dnsJSON, err := json.Marshal(dnsConfig)
	if err != nil {
		return 0, err
	}
	op.Dns = dnsJSON

	// 缓存策略
	op.CachePolicyId = cachePolicyId

	// WAF策略
	op.HttpFirewallPolicyId = httpFirewallPolicyId

	// 系统服务
	systemServicesJSON, err := json.Marshal(systemServices)
	if err != nil {
		return 0, err
	}
	op.SystemServices = systemServicesJSON

	op.UseAllAPINodes = 1
	op.ApiNodes = "[]"
	op.UniqueId = uniqueId
	op.Secret = secret
	op.State = NodeClusterStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// UpdateCluster 修改集群
func (this *NodeClusterDAO) UpdateCluster(tx *dbs.Tx, clusterId int64, name string, grantId int64, installDir string, timezone string) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	op := NewNodeClusterOperator()
	op.Id = clusterId
	op.Name = name
	op.GrantId = grantId
	op.InstallDir = installDir
	op.TimeZone = timezone
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// CountAllEnabledClusters 计算所有集群数量
func (this *NodeClusterDAO) CountAllEnabledClusters(tx *dbs.Tx, keyword string) (int64, error) {
	query := this.Query(tx).
		State(NodeClusterStateEnabled)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR dnsName like :keyword OR (dnsDomainId > 0 AND dnsDomainId IN (SELECT id FROM "+dns.SharedDNSDomainDAO.Table+" WHERE name LIKE :keyword AND state=1)))").
			Param("keyword", "%"+keyword+"%")
	}
	return query.Count()
}

// ListEnabledClusters 列出单页集群
func (this *NodeClusterDAO) ListEnabledClusters(tx *dbs.Tx, keyword string, offset, size int64) (result []*NodeCluster, err error) {
	query := this.Query(tx).
		State(NodeClusterStateEnabled)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR dnsName like :keyword OR (dnsDomainId > 0 AND dnsDomainId IN (SELECT id FROM "+dns.SharedDNSDomainDAO.Table+" WHERE name LIKE :keyword AND state=1)))").
			Param("keyword", "%"+keyword+"%")
	}
	_, err = query.
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// FindAllAPINodeAddrsWithCluster 查找所有API节点地址
func (this *NodeClusterDAO) FindAllAPINodeAddrsWithCluster(tx *dbs.Tx, clusterId int64) (result []string, err error) {
	one, err := this.Query(tx).
		Pk(clusterId).
		Result("useAllAPINodes", "apiNodes").
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	cluster := one.(*NodeCluster)
	if cluster.UseAllAPINodes == 1 {
		apiNodes, err := SharedAPINodeDAO.FindAllEnabledAPINodes(tx)
		if err != nil {
			return nil, err
		}
		for _, apiNode := range apiNodes {
			if apiNode.IsOn != 1 {
				continue
			}
			addrs, err := apiNode.DecodeAccessAddrStrings()
			if err != nil {
				return nil, err
			}
			result = append(result, addrs...)
		}
		return result, nil
	}

	apiNodeIds := []int64{}
	if !IsNotNull(cluster.ApiNodes) {
		return
	}
	err = json.Unmarshal([]byte(cluster.ApiNodes), &apiNodeIds)
	if err != nil {
		return nil, err
	}
	for _, apiNodeId := range apiNodeIds {
		apiNode, err := SharedAPINodeDAO.FindEnabledAPINode(tx, apiNodeId)
		if err != nil {
			return nil, err
		}
		if apiNode == nil || apiNode.IsOn != 1 {
			continue
		}
		addrs, err := apiNode.DecodeAccessAddrStrings()
		if err != nil {
			return nil, err
		}
		result = append(result, addrs...)
	}
	return result, nil
}

// FindClusterHealthCheckConfig 查找健康检查设置
func (this *NodeClusterDAO) FindClusterHealthCheckConfig(tx *dbs.Tx, clusterId int64) (*serverconfigs.HealthCheckConfig, error) {
	col, err := this.Query(tx).
		Pk(clusterId).
		Result("healthCheck").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(col) == 0 || col == "null" {
		return nil, nil
	}

	config := &serverconfigs.HealthCheckConfig{}
	err = json.Unmarshal([]byte(col), config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// UpdateClusterHealthCheck 修改健康检查设置
func (this *NodeClusterDAO) UpdateClusterHealthCheck(tx *dbs.Tx, clusterId int64, healthCheckJSON []byte) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId '" + strconv.FormatInt(clusterId, 10) + "'")
	}
	op := NewNodeClusterOperator()
	op.Id = clusterId
	op.HealthCheck = healthCheckJSON
	// 不需要通知更新
	return this.Save(tx, op)
}

// CountAllEnabledClustersWithGrantId 计算使用某个认证的集群数量
func (this *NodeClusterDAO) CountAllEnabledClustersWithGrantId(tx *dbs.Tx, grantId int64) (int64, error) {
	return this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("grantId", grantId).
		Count()
}

// FindAllEnabledClustersWithGrantId 获取使用某个认证的所有集群
func (this *NodeClusterDAO) FindAllEnabledClustersWithGrantId(tx *dbs.Tx, grantId int64) (result []*NodeCluster, err error) {
	_, err = this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("grantId", grantId).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// CountAllEnabledClustersWithDNSProviderId 计算使用某个DNS服务商的集群数量
func (this *NodeClusterDAO) CountAllEnabledClustersWithDNSProviderId(tx *dbs.Tx, dnsProviderId int64) (int64, error) {
	return this.Query(tx).
		State(NodeClusterStateEnabled).
		Where("dnsDomainId IN (SELECT id FROM "+dns.SharedDNSDomainDAO.Table+" WHERE state=1 AND providerId=:providerId)").
		Param("providerId", dnsProviderId).
		Count()
}

// FindAllEnabledClustersWithDNSProviderId 获取所有使用某个DNS服务商的集群
func (this *NodeClusterDAO) FindAllEnabledClustersWithDNSProviderId(tx *dbs.Tx, dnsProviderId int64) (result []*NodeCluster, err error) {
	_, err = this.Query(tx).
		State(NodeClusterStateEnabled).
		Where("dnsDomainId IN (SELECT id FROM "+dns.SharedDNSDomainDAO.Table+" WHERE state=1 AND providerId=:providerId)").
		Param("providerId", dnsProviderId).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// CountAllEnabledClustersWithDNSDomainId 计算使用某个DNS域名的集群数量
func (this *NodeClusterDAO) CountAllEnabledClustersWithDNSDomainId(tx *dbs.Tx, dnsDomainId int64) (int64, error) {
	return this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("dnsDomainId", dnsDomainId).
		Count()
}

// FindAllEnabledClusterIdsWithDNSDomainId 查询使用某个DNS域名的集群ID列表
func (this *NodeClusterDAO) FindAllEnabledClusterIdsWithDNSDomainId(tx *dbs.Tx, dnsDomainId int64) ([]int64, error) {
	ones, err := this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("dnsDomainId", dnsDomainId).
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*NodeCluster).Id))
	}
	return result, nil
}

// FindAllEnabledClustersWithDNSDomainId 查询使用某个DNS域名的所有集群域名
func (this *NodeClusterDAO) FindAllEnabledClustersWithDNSDomainId(tx *dbs.Tx, dnsDomainId int64) (result []*NodeCluster, err error) {
	_, err = this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("dnsDomainId", dnsDomainId).
		Result("id", "name", "dnsName", "dnsDomainId", "isOn").
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledClustersHaveDNSDomain 查询已经设置了域名的集群
func (this *NodeClusterDAO) FindAllEnabledClustersHaveDNSDomain(tx *dbs.Tx) (result []*NodeCluster, err error) {
	_, err = this.Query(tx).
		State(NodeClusterStateEnabled).
		Gt("dnsDomainId", 0).
		Result("id", "name", "dnsName", "dnsDomainId", "isOn").
		Slice(&result).
		FindAll()
	return
}

// FindClusterGrantId 查找集群的认证ID
func (this *NodeClusterDAO) FindClusterGrantId(tx *dbs.Tx, clusterId int64) (int64, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("grantId").
		FindInt64Col(0)
}

// FindClusterDNSInfo 查找DNS信息
func (this *NodeClusterDAO) FindClusterDNSInfo(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (*NodeCluster, error) {
	var cacheKey = this.Table + ":FindClusterDNSInfo:" + types.String(clusterId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(*NodeCluster), nil
		}
	}

	one, err := this.Query(tx).
		Pk(clusterId).
		Result("id", "name", "dnsName", "dnsDomainId", "dns", "isOn").
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	if cacheMap != nil {
		cacheMap.Put(cacheKey, one)
	}
	return one.(*NodeCluster), nil
}

// ExistClusterDNSName 检查某个子域名是否可用
func (this *NodeClusterDAO) ExistClusterDNSName(tx *dbs.Tx, dnsName string, excludeClusterId int64) (bool, error) {
	return this.Query(tx).
		Attr("dnsName", dnsName).
		State(NodeClusterStateEnabled).
		Where("id!=:clusterId").
		Param("clusterId", excludeClusterId).
		Exist()
}

// UpdateClusterDNS 修改集群DNS相关信息
func (this *NodeClusterDAO) UpdateClusterDNS(tx *dbs.Tx, clusterId int64, dnsName string, dnsDomainId int64, nodesAutoSync bool, serversAutoSync bool, cnameRecords []string, ttl int32) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	op := NewNodeClusterOperator()
	op.Id = clusterId
	op.DnsName = dnsName
	op.DnsDomainId = dnsDomainId

	if len(cnameRecords) == 0 {
		cnameRecords = []string{}
	}

	dnsConfig := &dnsconfigs.ClusterDNSConfig{
		NodesAutoSync:   nodesAutoSync,
		ServersAutoSync: serversAutoSync,
		CNameRecords:    cnameRecords,
		TTL:             ttl,
	}
	dnsJSON, err := json.Marshal(dnsConfig)
	if err != nil {
		return err
	}
	op.Dns = dnsJSON

	err = this.Save(tx, op)
	if err != nil {
		return err
	}
	err = this.NotifyUpdate(tx, clusterId)
	if err != nil {
		return err
	}
	return this.NotifyDNSUpdate(tx, clusterId)
}

// FindClusterAdminId 查找集群所属管理员
func (this *NodeClusterDAO) FindClusterAdminId(tx *dbs.Tx, clusterId int64) (int64, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("adminId").
		FindInt64Col(0)
}

// FindClusterTOAConfig 查找集群的TOA设置
func (this *NodeClusterDAO) FindClusterTOAConfig(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (*nodeconfigs.TOAConfig, error) {
	var cacheKey = this.Table + ":FindClusterTOAConfig:" + types.String(clusterId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(*nodeconfigs.TOAConfig), nil
		}
	}

	toa, err := this.Query(tx).
		Pk(clusterId).
		Result("toa").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if !IsNotNull(toa) {
		return nodeconfigs.DefaultTOAConfig(), nil
	}

	config := &nodeconfigs.TOAConfig{}
	err = json.Unmarshal([]byte(toa), config)
	if err != nil {
		return nil, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// UpdateClusterTOA 修改集群的TOA设置
func (this *NodeClusterDAO) UpdateClusterTOA(tx *dbs.Tx, clusterId int64, toaJSON []byte) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	op := NewNodeClusterOperator()
	op.Id = clusterId
	op.Toa = toaJSON
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// CountAllEnabledNodeClustersWithHTTPCachePolicyId 计算使用某个缓存策略的集群数量
func (this *NodeClusterDAO) CountAllEnabledNodeClustersWithHTTPCachePolicyId(tx *dbs.Tx, httpCachePolicyId int64) (int64, error) {
	return this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("cachePolicyId", httpCachePolicyId).
		Count()
}

// FindAllEnabledNodeClustersWithHTTPCachePolicyId 查找使用缓存策略的所有集群
func (this *NodeClusterDAO) FindAllEnabledNodeClustersWithHTTPCachePolicyId(tx *dbs.Tx, httpCachePolicyId int64) (result []*NodeCluster, err error) {
	_, err = this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("cachePolicyId", httpCachePolicyId).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledNodeClustersWithHTTPFirewallPolicyId 计算使用某个WAF策略的集群数量
func (this *NodeClusterDAO) CountAllEnabledNodeClustersWithHTTPFirewallPolicyId(tx *dbs.Tx, httpFirewallPolicyId int64) (int64, error) {
	return this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("httpFirewallPolicyId", httpFirewallPolicyId).
		Count()
}

// FindAllEnabledNodeClustersWithHTTPFirewallPolicyId 查找使用WAF策略的所有集群
func (this *NodeClusterDAO) FindAllEnabledNodeClustersWithHTTPFirewallPolicyId(tx *dbs.Tx, httpFirewallPolicyId int64) (result []*NodeCluster, err error) {
	_, err = this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("httpFirewallPolicyId", httpFirewallPolicyId).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledNodeClusterIdsWithHTTPFirewallPolicyId 查找使用WAF策略的所有集群Ids
func (this *NodeClusterDAO) FindAllEnabledNodeClusterIdsWithHTTPFirewallPolicyId(tx *dbs.Tx, httpFirewallPolicyId int64) (result []int64, err error) {
	ones, err := this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("httpFirewallPolicyId", httpFirewallPolicyId).
		ResultPk().
		FindAll()
	for _, one := range ones {
		result = append(result, int64(one.(*NodeCluster).Id))
	}
	return
}

// FindAllEnabledNodeClusterIdsWithCachePolicyId 查找使用缓存策略的所有集群Ids
func (this *NodeClusterDAO) FindAllEnabledNodeClusterIdsWithCachePolicyId(tx *dbs.Tx, cachePolicyId int64) (result []int64, err error) {
	ones, err := this.Query(tx).
		State(NodeClusterStateEnabled).
		Attr("cachePolicyId", cachePolicyId).
		ResultPk().
		FindAll()
	for _, one := range ones {
		result = append(result, int64(one.(*NodeCluster).Id))
	}
	return
}

// FindClusterHTTPFirewallPolicyId 获取集群的WAF策略ID
func (this *NodeClusterDAO) FindClusterHTTPFirewallPolicyId(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (int64, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":FindClusterHTTPFirewallPolicyId:" + types.String(clusterId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(int64), nil
	}

	firewallPolicyId, err := this.Query(tx).
		Pk(clusterId).
		Result("httpFirewallPolicyId").
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, firewallPolicyId)
	}

	return firewallPolicyId, nil
}

// UpdateNodeClusterHTTPCachePolicyId 设置集群的缓存策略
func (this *NodeClusterDAO) UpdateNodeClusterHTTPCachePolicyId(tx *dbs.Tx, clusterId int64, httpCachePolicyId int64) error {
	_, err := this.Query(tx).
		Pk(clusterId).
		Set("cachePolicyId", httpCachePolicyId).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterHTTPCachePolicyId 获取集群的缓存策略ID
func (this *NodeClusterDAO) FindClusterHTTPCachePolicyId(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (int64, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":FindClusterHTTPCachePolicyId:" + types.String(clusterId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(int64), nil
	}

	cachePolicyId, err := this.Query(tx).
		Pk(clusterId).
		Result("cachePolicyId").
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, cachePolicyId)
	}

	return cachePolicyId, nil
}

// UpdateNodeClusterHTTPFirewallPolicyId 设置集群的WAF策略
func (this *NodeClusterDAO) UpdateNodeClusterHTTPFirewallPolicyId(tx *dbs.Tx, clusterId int64, httpFirewallPolicyId int64) error {
	_, err := this.Query(tx).
		Pk(clusterId).
		Set("httpFirewallPolicyId", httpFirewallPolicyId).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// UpdateNodeClusterSystemService 修改集群的系统服务设置
func (this *NodeClusterDAO) UpdateNodeClusterSystemService(tx *dbs.Tx, clusterId int64, serviceType nodeconfigs.SystemServiceType, params maps.Map) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	service, err := this.Query(tx).
		Pk(clusterId).
		Result("systemServices").
		FindStringCol("")
	if err != nil {
		return err
	}
	servicesMap := map[string]maps.Map{}
	if IsNotNull(service) {
		err = json.Unmarshal([]byte(service), &servicesMap)
		if err != nil {
			return err
		}
	}

	if params == nil {
		params = maps.Map{}
	}
	servicesMap[serviceType] = params
	servicesJSON, err := json.Marshal(servicesMap)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(clusterId).
		Set("systemServices", servicesJSON).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindNodeClusterSystemServiceParams 查找集群的系统服务设置
func (this *NodeClusterDAO) FindNodeClusterSystemServiceParams(tx *dbs.Tx, clusterId int64, serviceType nodeconfigs.SystemServiceType) (params maps.Map, err error) {
	if clusterId <= 0 {
		return nil, errors.New("invalid clusterId")
	}
	service, err := this.Query(tx).
		Pk(clusterId).
		Result("systemServices").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	servicesMap := map[string]maps.Map{}
	if IsNotNull(service) {
		err = json.Unmarshal([]byte(service), &servicesMap)
		if err != nil {
			return nil, err
		}
	}
	return servicesMap[serviceType], nil
}

// FindNodeClusterSystemServices 查找集群的所有服务设置
func (this *NodeClusterDAO) FindNodeClusterSystemServices(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (services map[string]maps.Map, err error) {
	if clusterId <= 0 {
		return nil, errors.New("invalid clusterId")
	}

	var cacheKey = this.Table + ":FindNodeClusterSystemServices:" + types.String(clusterId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(map[string]maps.Map), nil
		}
	}

	service, err := this.Query(tx).
		Pk(clusterId).
		Result("systemServices").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	servicesMap := map[string]maps.Map{}
	if IsNotNull(service) {
		err = json.Unmarshal([]byte(service), &servicesMap)
		if err != nil {
			return nil, err
		}
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, servicesMap)
	}

	return servicesMap, nil
}

// GenUniqueId 生成唯一ID
func (this *NodeClusterDAO) GenUniqueId(tx *dbs.Tx) (string, error) {
	for {
		uniqueId := rands.HexString(32)
		ok, err := this.Query(tx).
			Attr("uniqueId", uniqueId).
			Exist()
		if err != nil {
			return "", err
		}
		if ok {
			continue
		}
		return uniqueId, nil
	}
}

// FindLatestNodeClusters 查询最近访问的集群
func (this *NodeClusterDAO) FindLatestNodeClusters(tx *dbs.Tx, size int64) (result []*NodeCluster, err error) {
	itemTable := SharedLatestItemDAO.Table
	itemType := LatestItemTypeCluster
	_, err = this.Query(tx).
		Result(this.Table+".id", this.Table+".name").
		Join(SharedLatestItemDAO, dbs.QueryJoinRight, this.Table+".id="+itemTable+".itemId AND "+itemTable+".itemType='"+itemType+"'").
		Asc("CEIL((UNIX_TIMESTAMP() - " + itemTable + ".updatedAt) / (7 * 86400))"). // 优先一个星期以内的
		Desc(itemTable + ".count").
		State(NodeClusterStateEnabled).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// CheckNodeClusterIsOn 获取集群是否正在启用状态
func (this *NodeClusterDAO) CheckNodeClusterIsOn(tx *dbs.Tx, clusterId int64) (bool, error) {
	return this.Query(tx).
		Pk(clusterId).
		State(NodeClusterStateEnabled).
		Attr("isOn", true).
		Exist()
}

// FindEnabledNodeClustersWithIds 查找一组集群
func (this *NodeClusterDAO) FindEnabledNodeClustersWithIds(tx *dbs.Tx, clusterIds []int64) (result []*NodeCluster, err error) {
	if len(clusterIds) == 0 {
		return
	}
	for _, clusterId := range clusterIds {
		cluster, err := this.Query(tx).
			Pk(clusterId).
			State(NodeClusterStateEnabled).
			Find()
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			continue
		}
		result = append(result, cluster.(*NodeCluster))
	}
	return
}

// ExistsEnabledCluster 检查集群是否存在
func (this *NodeClusterDAO) ExistsEnabledCluster(tx *dbs.Tx, clusterId int64) (bool, error) {
	if clusterId <= 0 {
		return false, nil
	}
	return this.Query(tx).
		Pk(clusterId).
		State(NodeClusterStateEnabled).
		Exist()
}

// FindClusterTimezone 查找时区
func (this *NodeClusterDAO) FindClusterTimezone(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (string, error) {
	var cacheKey = this.Table + ":FindEnabledTimeZone:" + types.String(clusterId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(string), nil
		}
	}

	timeZone, err := this.Query(tx).
		Pk(clusterId).
		Result("timeZone").
		FindStringCol("")
	if err != nil {
		return "", err
	}
	if cacheMap != nil {
		cacheMap.Put(cacheKey, timeZone)
	}
	return timeZone, nil
}

// NotifyUpdate 通知更新
func (this *NodeClusterDAO) NotifyUpdate(tx *dbs.Tx, clusterId int64) error {
	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, NodeTaskTypeConfigChanged)
}

// NotifyDNSUpdate 通知DNS更新
// TODO 更新新的DNS解析记录的同时，需要删除老的DNS解析记录
func (this *NodeClusterDAO) NotifyDNSUpdate(tx *dbs.Tx, clusterId int64) error {
	err := dns.SharedDNSTaskDAO.CreateClusterTask(tx, clusterId, dns.DNSTaskTypeClusterChange)
	if err != nil {
		return err
	}
	return nil
}
