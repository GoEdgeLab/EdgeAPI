package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
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
func (this *NodeClusterDAO) EnableNodeCluster(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeClusterStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeClusterDAO) DisableNodeCluster(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeClusterStateDisabled).
		Update()
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
