package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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

var SharedNodeClusterDAO = NewNodeClusterDAO()

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

// 查找启用中的条目
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
	op := NewNodeClusterOperator()
	op.Name = name
	op.GrantId = grantId
	op.InstallDir = installDir
	op.UseAllAPINodes = 1
	op.ApiNodes = "[]"
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
		FindAll()
	return
}
