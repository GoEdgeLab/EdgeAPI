package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodePriceItemStateEnabled  = 1 // 已启用
	NodePriceItemStateDisabled = 0 // 已禁用

	NodePriceTypeTraffic = "traffic" // 价格类型之流量
)

type NodePriceItemDAO dbs.DAO

func NewNodePriceItemDAO() *NodePriceItemDAO {
	return dbs.NewDAO(&NodePriceItemDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodePriceItems",
			Model:  new(NodePriceItem),
			PkName: "id",
		},
	}).(*NodePriceItemDAO)
}

var SharedNodePriceItemDAO *NodePriceItemDAO

func init() {
	dbs.OnReady(func() {
		SharedNodePriceItemDAO = NewNodePriceItemDAO()
	})
}

// EnableNodePriceItem 启用条目
func (this *NodePriceItemDAO) EnableNodePriceItem(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodePriceItemStateEnabled).
		Update()
	return err
}

// DisableNodePriceItem 禁用条目
func (this *NodePriceItemDAO) DisableNodePriceItem(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodePriceItemStateDisabled).
		Update()
	return err
}

// FindEnabledNodePriceItem 查找启用中的条目
func (this *NodePriceItemDAO) FindEnabledNodePriceItem(tx *dbs.Tx, id int64) (*NodePriceItem, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodePriceItemStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodePriceItem), err
}

// FindNodePriceItemName 根据主键查找名称
func (this *NodePriceItemDAO) FindNodePriceItemName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateItem 创建价格
func (this *NodePriceItemDAO) CreateItem(tx *dbs.Tx, name string, itemType string, bitsFrom, bitsTo int64) (int64, error) {
	op := NewNodePriceItemOperator()
	op.Name = name
	op.Type = itemType
	op.BitsFrom = bitsFrom
	op.BitsTo = bitsTo
	op.IsOn = true
	op.State = NodePriceItemStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateItem 修改价格
func (this *NodePriceItemDAO) UpdateItem(tx *dbs.Tx, itemId int64, name string, bitsFrom, bitsTo int64) error {
	if itemId <= 0 {
		return errors.New("invalid itemId")
	}
	op := NewNodePriceItemOperator()
	op.Id = itemId
	op.Name = name
	op.BitsFrom = bitsFrom
	op.BitsTo = bitsTo
	return this.Save(tx, op)
}

// FindAllEnabledRegionPrices 列出某个区域的所有价格
func (this *NodePriceItemDAO) FindAllEnabledRegionPrices(tx *dbs.Tx, priceType string) (result []*NodePriceItem, err error) {
	_, err = this.Query(tx).
		Attr("type", priceType).
		State(NodePriceItemStateEnabled).
		Asc("bitsFrom").
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledAndOnRegionPrices 列出某个区域的所有启用的价格
func (this *NodePriceItemDAO) FindAllEnabledAndOnRegionPrices(tx *dbs.Tx, priceType string) (result []*NodePriceItem, err error) {
	_, err = this.Query(tx).
		Attr("type", priceType).
		State(NodePriceItemStateEnabled).
		Attr("isOn", true).
		Asc("bitsFrom").
		Slice(&result).
		FindAll()
	return
}

// SearchItemsWithBytes 根据字节查找付费项目
func (this *NodePriceItemDAO) SearchItemsWithBytes(items []*NodePriceItem, bytes int64) int64 {
	bytes *= 8

	for _, item := range items {
		if bytes >= int64(item.BitsFrom) && (bytes < int64(item.BitsTo) || item.BitsTo == 0) {
			return int64(item.Id)
		}
	}
	return 0
}
