package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
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
func (this *NodeClusterDAO) DisableNodeCluster(tx *dbs.Tx, clusterId int64) error {
	_, err := this.Query(tx).
		Pk(clusterId).
		Set("state", NodeClusterStateDisabled).
		Update()
	if err != nil {
		return err
	}

	return SharedNodeLogDAO.DeleteNodeLogsWithCluster(tx, nodeconfigs.NodeRoleNode, clusterId)
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
		Desc("isPinned").
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
func (this *NodeClusterDAO) CreateCluster(tx *dbs.Tx, adminId int64, name string, grantId int64, installDir string, dnsDomainId int64, dnsName string, dnsTTL int32, cachePolicyId int64, httpFirewallPolicyId int64, systemServices map[string]maps.Map, globalServerConfig *serverconfigs.GlobalServerConfig, autoInstallNftables bool) (clusterId int64, err error) {
	uniqueId, err := this.GenUniqueId(tx)
	if err != nil {
		return 0, err
	}

	var secret = rands.String(32)
	err = SharedApiTokenDAO.CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleCluster)
	if err != nil {
		return 0, err
	}

	var op = NewNodeClusterOperator()
	op.AdminId = adminId
	op.Name = name
	op.GrantId = grantId
	op.InstallDir = installDir

	// DNS设置
	op.DnsDomainId = dnsDomainId
	op.DnsName = dnsName
	var dnsConfig = &dnsconfigs.ClusterDNSConfig{
		NodesAutoSync:    true,
		ServersAutoSync:  true,
		CNAMERecords:     []string{},
		CNAMEAsDomain:    true,
		TTL:              dnsTTL,
		IncludingLnNodes: true,
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

	// 全局服务配置
	if globalServerConfig == nil {
		globalServerConfig = serverconfigs.DefaultGlobalServerConfig()
	}
	globalServerConfigJSON, err := json.Marshal(globalServerConfig)
	if err != nil {
		return 0, err
	}
	op.GlobalServerConfig = globalServerConfigJSON

	op.UseAllAPINodes = 1
	op.ApiNodes = "[]"
	op.UniqueId = uniqueId
	op.Secret = secret
	op.AutoInstallNftables = autoInstallNftables
	op.State = NodeClusterStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// UpdateCluster 修改集群
func (this *NodeClusterDAO) UpdateCluster(tx *dbs.Tx, clusterId int64, name string, grantId int64, installDir string, timezone string, nodeMaxThreads int32, autoOpenPorts bool, clockConfig *nodeconfigs.ClockConfig, autoRemoteStart bool, autoInstallTables bool, sshParams *nodeconfigs.SSHParams) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	var op = NewNodeClusterOperator()
	op.Id = clusterId
	op.Name = name
	op.GrantId = grantId
	op.InstallDir = installDir
	op.TimeZone = timezone

	if nodeMaxThreads < 0 {
		nodeMaxThreads = 0
	}
	op.NodeMaxThreads = nodeMaxThreads
	op.AutoOpenPorts = autoOpenPorts

	if clockConfig != nil {
		clockJSON, err := json.Marshal(clockConfig)
		if err != nil {
			return err
		}
		op.Clock = clockJSON
	}

	op.AutoRemoteStart = autoRemoteStart
	op.AutoInstallNftables = autoInstallTables

	if sshParams != nil {
		sshParamsJSON, err := json.Marshal(sshParams)
		if err != nil {
			return err
		}
		op.SshParams = sshParamsJSON
	}

	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// UpdateClusterIsPinned 设置集群是否置顶
func (this *NodeClusterDAO) UpdateClusterIsPinned(tx *dbs.Tx, clusterId int64, isPinned bool) error {
	return this.Query(tx).
		Pk(clusterId).
		Set("isPinned", isPinned).
		UpdateQuickly()
}

// CountAllEnabledClusters 计算所有集群数量
func (this *NodeClusterDAO) CountAllEnabledClusters(tx *dbs.Tx, keyword string) (int64, error) {
	query := this.Query(tx).
		State(NodeClusterStateEnabled)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR dnsName like :keyword OR (dnsDomainId > 0 AND dnsDomainId IN (SELECT id FROM "+dns.SharedDNSDomainDAO.Table+" WHERE name LIKE :keyword AND state=1)))").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	return query.Count()
}

// ListEnabledClusters 列出单页集群
func (this *NodeClusterDAO) ListEnabledClusters(tx *dbs.Tx, keyword string, offset, size int64) (result []*NodeCluster, err error) {
	query := this.Query(tx).
		State(NodeClusterStateEnabled)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR dnsName like :keyword OR (dnsDomainId > 0 AND dnsDomainId IN (SELECT id FROM "+dns.SharedDNSDomainDAO.Table+" WHERE name LIKE :keyword AND state=1)))").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	_, err = query.
		Offset(offset).
		Limit(size).
		Slice(&result).
		Desc("isPinned").
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
			if !apiNode.IsOn {
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
	err = json.Unmarshal(cluster.ApiNodes, &apiNodeIds)
	if err != nil {
		return nil, err
	}
	for _, apiNodeId := range apiNodeIds {
		apiNode, err := SharedAPINodeDAO.FindEnabledAPINode(tx, apiNodeId, nil)
		if err != nil {
			return nil, err
		}
		if apiNode == nil || !apiNode.IsOn {
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
	var op = NewNodeClusterOperator()
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
		Result("id", "name", "dnsName", "dnsDomainId", "isOn", "dns").
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

// FindClusterSSHParams 查找集群的SSH默认参数
func (this *NodeClusterDAO) FindClusterSSHParams(tx *dbs.Tx, clusterId int64) (*nodeconfigs.SSHParams, error) {
	sshParamsJSON, err := this.Query(tx).
		Pk(clusterId).
		Result("sshParams").
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	var params = nodeconfigs.DefaultSSHParams()
	if len(sshParamsJSON) == 0 {
		return params, nil
	}
	err = json.Unmarshal(sshParamsJSON, params)
	if err != nil {
		return nil, err
	}
	return params, nil
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
		Result("id", "name", "dnsName", "dnsDomainId", "dns", "isOn", "state").
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
func (this *NodeClusterDAO) UpdateClusterDNS(tx *dbs.Tx, clusterId int64, dnsName string, dnsDomainId int64, nodesAutoSync bool, serversAutoSync bool, cnameRecords []string, ttl int32, cnameAsDomain bool, includingLnNodes bool) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}

	// 删除老的域名中相关记录
	oldOne, err := this.Query(tx).
		Pk(clusterId).
		Result("dnsName", "dnsDomainId").
		Find()
	if err != nil {
		return err
	}
	if oldOne == nil {
		return nil
	}

	var oldCluster = oldOne.(*NodeCluster)
	var oldDNSDomainId = int64(oldCluster.DnsDomainId)
	var shouldRemoveOld = false
	if (oldDNSDomainId > 0 && oldDNSDomainId != dnsDomainId) || (oldCluster.DnsName != dnsName) {
		if oldDNSDomainId == dnsDomainId {
			// 如果只是换子域名，需要在新的域名添加之前，先删除老的子域名，防止无法添加CNAME
			err = dns.SharedDNSTaskDAO.CreateClusterRemoveTask(tx, clusterId, oldDNSDomainId, oldCluster.DnsName)
			if err != nil {
				return err
			}
		} else {
			shouldRemoveOld = true
		}
	}

	var op = NewNodeClusterOperator()
	op.Id = clusterId
	op.DnsName = dnsName
	op.DnsDomainId = dnsDomainId

	if len(cnameRecords) == 0 {
		cnameRecords = []string{}
	}

	var dnsConfig = &dnsconfigs.ClusterDNSConfig{
		NodesAutoSync:    nodesAutoSync,
		ServersAutoSync:  serversAutoSync,
		CNAMERecords:     cnameRecords,
		TTL:              ttl,
		CNAMEAsDomain:    cnameAsDomain,
		IncludingLnNodes: includingLnNodes,
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
	err = this.NotifyDNSUpdate(tx, clusterId)
	if err != nil {
		return err
	}

	// 删除老的记录
	if shouldRemoveOld {
		err = dns.SharedDNSTaskDAO.CreateClusterRemoveTask(tx, clusterId, oldDNSDomainId, oldCluster.DnsName)
		if err != nil {
			return err
		}
	}

	return nil
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
	if !IsNotNull([]byte(toa)) {
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
	var op = NewNodeClusterOperator()
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

// FindAllEnabledNodeClusterIds 查找所有可用的集群
func (this *NodeClusterDAO) FindAllEnabledNodeClusterIds(tx *dbs.Tx) ([]int64, error) {
	ones, err := this.Query(tx).
		State(NodeClusterStateEnabled).
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	var result = []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*NodeCluster).Id))
	}
	return result, nil
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
	serviceData, err := this.Query(tx).
		Pk(clusterId).
		Result("systemServices").
		FindBytesCol()
	if err != nil {
		return err
	}
	servicesMap := map[string]maps.Map{}
	if IsNotNull(serviceData) {
		err = json.Unmarshal(serviceData, &servicesMap)
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
	if IsNotNull([]byte(service)) {
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
	if IsNotNull([]byte(service)) {
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

// FindClusterBasicInfo 查找集群基础信息
func (this *NodeClusterDAO) FindClusterBasicInfo(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (*NodeCluster, error) {
	var cacheKey = this.Table + ":FindClusterBasicInfo:" + types.String(clusterId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(*NodeCluster), nil
		}
	}

	cluster, err := this.Query(tx).
		Pk(clusterId).
		State(NodeClusterStateEnabled).
		Result("id", "name", "timeZone", "nodeMaxThreads", "cachePolicyId", "httpFirewallPolicyId", "autoOpenPorts", "webp", "uam", "isOn", "ddosProtection", "clock", "globalServerConfig", "autoInstallNftables").
		Find()
	if err != nil || cluster == nil {
		return nil, err
	}
	if cacheMap != nil {
		cacheMap.Put(cacheKey, cluster)
	}
	return cluster.(*NodeCluster), nil
}

// UpdateClusterWebPPolicy 修改WebP设置
func (this *NodeClusterDAO) UpdateClusterWebPPolicy(tx *dbs.Tx, clusterId int64, webpPolicy *nodeconfigs.WebPImagePolicy) error {
	if webpPolicy == nil {
		err := this.Query(tx).
			Pk(clusterId).
			Set("webp", dbs.SQL("null")).
			UpdateQuickly()
		if err != nil {
			return err
		}

		return this.NotifyUpdate(tx, clusterId)
	}

	webpPolicyJSON, err := json.Marshal(webpPolicy)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("webp", webpPolicyJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterWebPPolicy 查询WebP设置
func (this *NodeClusterDAO) FindClusterWebPPolicy(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (*nodeconfigs.WebPImagePolicy, error) {
	var cacheKey = this.Table + ":FindClusterWebPPolicy:" + types.String(clusterId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(*nodeconfigs.WebPImagePolicy), nil
		}
	}

	webpJSON, err := this.Query(tx).
		Pk(clusterId).
		Result("webp").
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	if IsNull(webpJSON) {
		return nodeconfigs.DefaultWebPImagePolicy, nil
	}

	var policy = &nodeconfigs.WebPImagePolicy{}
	err = json.Unmarshal(webpJSON, policy)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// UpdateClusterUAMPolicy 修改UAM设置
func (this *NodeClusterDAO) UpdateClusterUAMPolicy(tx *dbs.Tx, clusterId int64, uamPolicy *nodeconfigs.UAMPolicy) error {
	if uamPolicy == nil {
		err := this.Query(tx).
			Pk(clusterId).
			Set("uam", dbs.SQL("null")).
			UpdateQuickly()
		if err != nil {
			return err
		}

		return this.NotifyUAMUpdate(tx, clusterId)
	}

	uamPolicyJSON, err := json.Marshal(uamPolicy)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("uam", uamPolicyJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUAMUpdate(tx, clusterId)
}

// FindClusterUAMPolicy 查询设置
func (this *NodeClusterDAO) FindClusterUAMPolicy(tx *dbs.Tx, clusterId int64, cacheMap *utils.CacheMap) (*nodeconfigs.UAMPolicy, error) {
	var cacheKey = this.Table + ":FindClusterUAMPolicy:" + types.String(clusterId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(*nodeconfigs.UAMPolicy), nil
		}
	}

	uamJSON, err := this.Query(tx).
		Pk(clusterId).
		Result("uam").
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	if IsNull(uamJSON) {
		return nodeconfigs.DefaultUAMPolicy, nil
	}

	var policy = &nodeconfigs.UAMPolicy{}
	err = json.Unmarshal(uamJSON, policy)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// FindClusterDDoSProtection 获取集群的DDoS设置
func (this *NodeClusterDAO) FindClusterDDoSProtection(tx *dbs.Tx, clusterId int64) (*ddosconfigs.ProtectionConfig, error) {
	one, err := this.Query(tx).
		Result("ddosProtection").
		Pk(clusterId).
		Find()
	if one == nil || err != nil {
		return nil, err
	}

	return one.(*NodeCluster).DecodeDDoSProtection(), nil
}

// UpdateClusterDDoSProtection 设置集群的DDoS设置
func (this *NodeClusterDAO) UpdateClusterDDoSProtection(tx *dbs.Tx, clusterId int64, ddosProtection *ddosconfigs.ProtectionConfig) error {
	if clusterId <= 0 {
		return ErrNotFound
	}

	var op = NewNodeClusterOperator()
	op.Id = clusterId

	if ddosProtection == nil {
		op.DdosProtection = "{}"
	} else {
		ddosProtectionJSON, err := json.Marshal(ddosProtection)
		if err != nil {
			return err
		}
		op.DdosProtection = ddosProtectionJSON
	}

	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, 0, NodeTaskTypeDDosProtectionChanged)
}

// FindClusterGlobalServerConfig 查询全局服务配置
func (this *NodeClusterDAO) FindClusterGlobalServerConfig(tx *dbs.Tx, clusterId int64) (*serverconfigs.GlobalServerConfig, error) {
	configJSON, err := this.Query(tx).
		Pk(clusterId).
		Result("globalServerConfig").
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	var config = serverconfigs.DefaultGlobalServerConfig()
	if IsNull(configJSON) {
		return config, nil
	}

	err = json.Unmarshal(configJSON, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// UpdateClusterGlobalServerConfig 修改全局服务配置
func (this *NodeClusterDAO) UpdateClusterGlobalServerConfig(tx *dbs.Tx, clusterId int64, config *serverconfigs.GlobalServerConfig) error {
	if config == nil {
		config = serverconfigs.DefaultGlobalServerConfig()
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("globalServerConfig", configJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, 0, NodeTaskTypeGlobalServerConfigChanged)
}

// NotifyUpdate 通知更新
func (this *NodeClusterDAO) NotifyUpdate(tx *dbs.Tx, clusterId int64) error {
	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, 0, NodeTaskTypeConfigChanged)
}

// NotifyUAMUpdate 通知UAM更新
func (this *NodeClusterDAO) NotifyUAMUpdate(tx *dbs.Tx, clusterId int64) error {
	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, 0, NodeTaskTypeUAMPolicyChanged)
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
