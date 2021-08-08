package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodeClusterMetricItemStateEnabled  = 1 // 已启用
	NodeClusterMetricItemStateDisabled = 0 // 已禁用
)

type NodeClusterMetricItemDAO dbs.DAO

func NewNodeClusterMetricItemDAO() *NodeClusterMetricItemDAO {
	return dbs.NewDAO(&NodeClusterMetricItemDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeClusterMetricItems",
			Model:  new(NodeClusterMetricItem),
			PkName: "id",
		},
	}).(*NodeClusterMetricItemDAO)
}

var SharedNodeClusterMetricItemDAO *NodeClusterMetricItemDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeClusterMetricItemDAO = NewNodeClusterMetricItemDAO()
	})
}

// EnableNodeClusterMetricItem 启用条目
func (this *NodeClusterMetricItemDAO) EnableNodeClusterMetricItem(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeClusterMetricItemStateEnabled).
		Update()
	return err
}

// DisableNodeClusterMetricItem 禁用条目
func (this *NodeClusterMetricItemDAO) DisableNodeClusterMetricItem(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeClusterMetricItemStateDisabled).
		Update()
	return err
}

// FindEnabledNodeClusterMetricItem 查找启用中的条目
func (this *NodeClusterMetricItemDAO) FindEnabledNodeClusterMetricItem(tx *dbs.Tx, id uint32) (*NodeClusterMetricItem, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeClusterMetricItemStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeClusterMetricItem), err
}

// DisableClusterItem 禁用某个集群的指标
func (this *NodeClusterMetricItemDAO) DisableClusterItem(tx *dbs.Tx, clusterId int64, itemId int64) error {
	err := this.Query(tx).
		Attr("clusterId", clusterId).
		Attr("itemId", itemId).
		State(NodeClusterMetricItemStateEnabled).
		Set("state", NodeClusterMetricItemStateDisabled).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// EnableClusterItem 启用某个集群的指标
func (this *NodeClusterMetricItemDAO) EnableClusterItem(tx *dbs.Tx, clusterId int64, itemId int64) error {
	if clusterId <= 0 || itemId <= 0 {
		return errors.New("clusterId or itemId should not be 0")
	}
	var op = NewNodeClusterMetricItemOperator()
	op.ClusterId = clusterId
	op.ItemId = itemId
	op.IsOn = true
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindAllClusterItems 查找某个集群的指标
// category 不填写即表示获取所有指标
func (this *NodeClusterMetricItemDAO) FindAllClusterItems(tx *dbs.Tx, clusterId int64, category string) (result []*NodeClusterMetricItem, err error) {
	var query = this.Query(tx).
		Attr("clusterId", clusterId).
		State(NodeClusterMetricItemStateEnabled)
	if len(category) > 0 {
		query.Where("itemId IN (SELECT id FROM "+SharedMetricItemDAO.Table+" WHERE category=:category)").
			Param("category", category)
	}
	_, err = query.
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllClusterItemIds 查找某个集群的指标Ids
func (this *NodeClusterMetricItemDAO) FindAllClusterItemIds(tx *dbs.Tx, clusterId int64) (result []int64, err error) {
	ones, err := this.Query(tx).
		Attr("clusterId", clusterId).
		State(NodeClusterMetricItemStateEnabled).
		Result("itemId").
		DescPk().
		FindAll()
	for _, one := range ones {
		result = append(result, int64(one.(*NodeClusterMetricItem).ItemId))
	}
	return
}

// FindAllClusterIdsWithItemId 查找使用某个指标的所有集群IDs
func (this *NodeClusterMetricItemDAO) FindAllClusterIdsWithItemId(tx *dbs.Tx, itemId int64) (clusterIds []int64, err error) {
	ones, err := this.Query(tx).
		Attr("itemId", itemId).
		Result("clusterId").
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		clusterIds = append(clusterIds, int64(one.(*NodeClusterMetricItem).ClusterId))
	}
	return
}

// CountAllClusterItems 计算集群中指标数量
func (this *NodeClusterMetricItemDAO) CountAllClusterItems(tx *dbs.Tx, clusterId int64) (int64, error) {
	return this.Query(tx).
		Attr("clusterId", clusterId).
		State(NodeClusterMetricItemStateEnabled).
		Count()
}

// ExistsClusterItem 是否存在
func (this *NodeClusterMetricItemDAO) ExistsClusterItem(tx *dbs.Tx, clusterId int64, itemId int64) (bool, error) {
	return this.Query(tx).
		Attr("clusterId", clusterId).
		Attr("itemId", itemId).
		State(NodeClusterMetricItemStateEnabled).
		Exist()
}

// NotifyUpdate 通知更新
func (this *NodeClusterMetricItemDAO) NotifyUpdate(tx *dbs.Tx, clusterId int64) error {
	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, NodeTaskTypeConfigChanged)
}
