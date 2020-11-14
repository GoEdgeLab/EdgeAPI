package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
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

// 启用条目
func (this *NodeClusterDAO) EnableNodeCluster(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", NodeClusterStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *NodeClusterDAO) DisableNodeCluster(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", NodeClusterStateDisabled).
		Update()
	return err
}

// 查找集群
func (this *NodeClusterDAO) FindEnabledNodeCluster(id int64) (*NodeCluster, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", NodeClusterStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeCluster), err
}

// 根据UniqueId获取ID
// TODO 增加缓存
func (this *NodeClusterDAO) FindEnabledClusterIdWithUniqueId(uniqueId string) (int64, error) {
	return this.Query().
		State(NodeClusterStateEnabled).
		Attr("uniqueId", uniqueId).
		ResultPk().
		FindInt64Col(0)
}

// 根据主键查找名称
func (this *NodeClusterDAO) FindNodeClusterName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 查找所有可用的集群
func (this *NodeClusterDAO) FindAllEnableClusters() (result []*NodeCluster, err error) {
	_, err = this.Query().
		State(NodeClusterStateEnabled).
		Slice(&result).
		Desc("order").
		DescPk().
		FindAll()
	return
}

// 创建集群
func (this *NodeClusterDAO) CreateCluster(name string, grantId int64, installDir string) (clusterId int64, err error) {
	uniqueId, err := this.genUniqueId()
	if err != nil {
		return 0, err
	}

	secret := rands.String(32)
	err = SharedApiTokenDAO.CreateAPIToken(uniqueId, secret, NodeRoleCluster)
	if err != nil {
		return 0, err
	}

	op := NewNodeClusterOperator()
	op.Name = name
	op.GrantId = grantId
	op.InstallDir = installDir
	op.UseAllAPINodes = 1
	op.ApiNodes = "[]"
	op.UniqueId = uniqueId
	op.Secret = secret
	op.State = NodeClusterStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// 修改集群
func (this *NodeClusterDAO) UpdateCluster(clusterId int64, name string, grantId int64, installDir string) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	op := NewNodeClusterOperator()
	op.Id = clusterId
	op.Name = name
	op.GrantId = grantId
	op.InstallDir = installDir
	_, err := this.Save(op)
	return err
}

// 计算所有集群数量
func (this *NodeClusterDAO) CountAllEnabledClusters() (int64, error) {
	return this.Query().
		State(NodeClusterStateEnabled).
		Count()
}

// 列出单页集群
func (this *NodeClusterDAO) ListEnabledClusters(offset, size int64) (result []*NodeCluster, err error) {
	_, err = this.Query().
		State(NodeClusterStateEnabled).
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// 查找所有API节点地址
func (this *NodeClusterDAO) FindAllAPINodeAddrsWithCluster(clusterId int64) (result []string, err error) {
	one, err := this.Query().
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
		apiNodes, err := SharedAPINodeDAO.FindAllEnabledAPINodes()
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
		apiNode, err := SharedAPINodeDAO.FindEnabledAPINode(apiNodeId)
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

// 查找健康检查设置
func (this *NodeClusterDAO) FindClusterHealthCheckConfig(clusterId int64) (*serverconfigs.HealthCheckConfig, error) {
	col, err := this.Query().
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

// 修改健康检查设置
func (this *NodeClusterDAO) UpdateClusterHealthCheck(clusterId int64, healthCheckJSON []byte) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId '" + strconv.FormatInt(clusterId, 10) + "'")
	}
	op := NewNodeClusterOperator()
	op.Id = clusterId
	op.HealthCheck = healthCheckJSON
	_, err := this.Save(op)
	return err
}

// 计算使用某个认证的集群数量
func (this *NodeClusterDAO) CountAllEnabledClustersWithGrantId(grantId int64) (int64, error) {
	return this.Query().
		State(NodeClusterStateEnabled).
		Attr("grantId", grantId).
		Count()
}

// 获取使用某个认证的所有集群
func (this *NodeClusterDAO) FindAllEnabledClustersWithGrantId(grantId int64) (result []*NodeCluster, err error) {
	_, err = this.Query().
		State(NodeClusterStateEnabled).
		Attr("grantId", grantId).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// 计算使用某个DNS服务商的集群数量
func (this *NodeClusterDAO) CountAllEnabledClustersWithDNSProviderId(dnsProviderId int64) (int64, error) {
	return this.Query().
		State(NodeClusterStateEnabled).
		Where("dnsDomainId IN (SELECT id FROM "+SharedDNSDomainDAO.Table+" WHERE state=1 AND providerId=:providerId)").
		Param("providerId", dnsProviderId).
		Count()
}

// 获取所有使用某个DNS服务商的集群
func (this *NodeClusterDAO) FindAllEnabledClustersWithDNSProviderId(dnsProviderId int64) (result []*NodeCluster, err error) {
	_, err = this.Query().
		State(NodeClusterStateEnabled).
		Where("dnsDomainId IN (SELECT id FROM "+SharedDNSDomainDAO.Table+" WHERE state=1 AND providerId=:providerId)").
		Param("providerId", dnsProviderId).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// 计算使用某个DNS域名的集群数量
func (this *NodeClusterDAO) CountAllEnabledClustersWithDNSDomainId(dnsDomainId int64) (int64, error) {
	return this.Query().
		State(NodeClusterStateEnabled).
		Attr("dnsDomainId", dnsDomainId).
		Count()
}

// 查询使用某个DNS域名的集群ID列表
func (this *NodeClusterDAO) FindAllEnabledClusterIdsWithDNSDomainId(dnsDomainId int64) ([]int64, error) {
	ones, err := this.Query().
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

// 查询使用某个DNS域名的所有集群域名
func (this *NodeClusterDAO) FindAllEnabledClustersWithDNSDomainId(dnsDomainId int64) (result []*NodeCluster, err error) {
	_, err = this.Query().
		State(NodeClusterStateEnabled).
		Attr("dnsDomainId", dnsDomainId).
		Result("id", "name", "dnsName", "dnsDomainId").
		Slice(&result).
		FindAll()
	return
}

// 查询已经设置了域名的集群
func (this *NodeClusterDAO) FindAllEnabledClustersHaveDNSDomain() (result []*NodeCluster, err error) {
	_, err = this.Query().
		State(NodeClusterStateEnabled).
		Gt("dnsDomainId", 0).
		Result("id", "name", "dnsName", "dnsDomainId").
		Slice(&result).
		FindAll()
	return
}

// 查找集群的认证ID
func (this *NodeClusterDAO) FindClusterGrantId(clusterId int64) (int64, error) {
	return this.Query().
		Pk(clusterId).
		Result("grantId").
		FindInt64Col(0)
}

// 查找DNS信息
func (this *NodeClusterDAO) FindClusterDNSInfo(clusterId int64) (*NodeCluster, error) {
	one, err := this.Query().
		Pk(clusterId).
		Result("id", "name", "dnsName", "dnsDomainId").
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*NodeCluster), nil
}

// 检查某个子域名是否可用
func (this *NodeClusterDAO) ExistClusterDNSName(dnsName string, excludeClusterId int64) (bool, error) {
	return this.Query().
		Attr("dnsName", dnsName).
		State(NodeClusterStateEnabled).
		Where("id!=:clusterId").
		Param("clusterId", excludeClusterId).
		Exist()
}

// 修改集群DNS相关信息
func (this *NodeClusterDAO) UpdateClusterDNS(clusterId int64, dnsName string, dnsDomainId int64) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	op := NewNodeClusterOperator()
	op.Id = clusterId
	op.DnsName = dnsName
	op.DnsDomainId = dnsDomainId
	_, err := this.Save(op)
	return err
}

// 检查集群的DNS问题
func (this *NodeClusterDAO) CheckClusterDNS(cluster *NodeCluster) (issues []*pb.DNSIssue, err error) {
	clusterId := int64(cluster.Id)
	domainId := int64(cluster.DnsDomainId)

	// 检查域名
	domain, err := SharedDNSDomainDAO.FindEnabledDNSDomain(domainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		issues = append(issues, &pb.DNSIssue{
			Target:      cluster.Name,
			TargetId:    clusterId,
			Type:        "cluster",
			Description: "域名选择错误，需要重新选择",
			Params:      nil,
		})
		return
	}

	// 检查二级域名
	if len(cluster.DnsName) == 0 {
		issues = append(issues, &pb.DNSIssue{
			Target:      cluster.Name,
			TargetId:    clusterId,
			Type:        "cluster",
			Description: "没有设置二级域名",
			Params:      nil,
		})
		return
	}

	// TODO 检查域名格式

	// TODO 检查域名是否已解析

	// 检查节点
	nodes, err := SharedNodeDAO.FindAllEnabledNodesDNSWithClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	// TODO 检查节点数量不能为0

	for _, node := range nodes {
		nodeId := int64(node.Id)

		route, err := node.DNSRoute(domainId)
		if err != nil {
			return nil, err
		}
		if len(route) == 0 {
			issues = append(issues, &pb.DNSIssue{
				Target:      node.Name,
				TargetId:    nodeId,
				Type:        "node",
				Description: "没有选择节点所属线路",
				Params: map[string]string{
					"clusterName": cluster.Name,
					"clusterId":   numberutils.FormatInt64(clusterId),
				},
			})
			continue
		}

		// 检查线路是否在已有线路中
		routeOk, err := domain.ContainsRoute(route)
		if err != nil {
			return nil, err
		}
		if !routeOk {
			issues = append(issues, &pb.DNSIssue{
				Target:      node.Name,
				TargetId:    nodeId,
				Type:        "node",
				Description: "线路已经失效，请重新选择",
				Params: map[string]string{
					"clusterName": cluster.Name,
					"clusterId":   numberutils.FormatInt64(clusterId),
				},
			})
			continue
		}

		// 检查IP地址
		ipAddr, err := SharedNodeIPAddressDAO.FindFirstNodeIPAddress(nodeId)
		if err != nil {
			return nil, err
		}
		if len(ipAddr) == 0 {
			issues = append(issues, &pb.DNSIssue{
				Target:      node.Name,
				TargetId:    nodeId,
				Type:        "node",
				Description: "没有设置IP地址",
				Params: map[string]string{
					"clusterName": cluster.Name,
					"clusterId":   numberutils.FormatInt64(clusterId),
				},
			})
			continue
		}

		// TODO 检查是否有解析记录
	}

	return
}

// 生成唯一ID
func (this *NodeClusterDAO) genUniqueId() (string, error) {
	for {
		uniqueId := rands.HexString(32)
		ok, err := this.Query().
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
