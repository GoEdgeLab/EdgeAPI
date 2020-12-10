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

// 启用条目
func (this *NodePriceItemDAO) EnableNodePriceItem(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", NodePriceItemStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *NodePriceItemDAO) DisableNodePriceItem(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", NodePriceItemStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *NodePriceItemDAO) FindEnabledNodePriceItem(id int64) (*NodePriceItem, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", NodePriceItemStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodePriceItem), err
}

// 根据主键查找名称
func (this *NodePriceItemDAO) FindNodePriceItemName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建价格
func (this *NodePriceItemDAO) CreateItem(name string, itemType string, bitsFrom, bitsTo int64) (int64, error) {
	op := NewNodePriceItemOperator()
	op.Name = name
	op.Type = itemType
	op.BitsFrom = bitsFrom
	op.BitsTo = bitsTo
	op.IsOn = true
	op.State = NodePriceItemStateEnabled
	return this.SaveInt64(op)
}

// 修改价格
func (this *NodePriceItemDAO) UpdateItem(itemId int64, name string, bitsFrom, bitsTo int64) error {
	if itemId <= 0 {
		return errors.New("invalid itemId")
	}
	op := NewNodePriceItemOperator()
	op.Id = itemId
	op.Name = name
	op.BitsFrom = bitsFrom
	op.BitsTo = bitsTo
	return this.Save(op)
}

// 列出某个区域的所有价格
func (this *NodePriceItemDAO) FindAllEnabledRegionPrices(priceType string) (result []*NodePriceItem, err error) {
	_, err = this.Query().
		Attr("type", priceType).
		State(NodePriceItemStateEnabled).
		Asc("bitsFrom").
		Slice(&result).
		FindAll()
	return
}

// 列出某个区域的所有启用的价格
func (this *NodePriceItemDAO) FindAllEnabledAndOnRegionPrices(priceType string) (result []*NodePriceItem, err error) {
	_, err = this.Query().
		Attr("type", priceType).
		State(NodePriceItemStateEnabled).
		Attr("isOn", true).
		Asc("bitsFrom").
		Slice(&result).
		FindAll()
	return
}
