package models

import (
	"encoding/json"
	"errors"
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

// 查找集群的认证ID
func (this *NodeClusterDAO) FindClusterGrantId(clusterId int64) (int64, error) {
	return this.Query().
		Pk(clusterId).
		Result("grantId").
		FindInt64Col(0)
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
