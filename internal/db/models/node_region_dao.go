package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodeRegionStateEnabled  = 1 // 已启用
	NodeRegionStateDisabled = 0 // 已禁用
)

type NodeRegionDAO dbs.DAO

func NewNodeRegionDAO() *NodeRegionDAO {
	return dbs.NewDAO(&NodeRegionDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeRegions",
			Model:  new(NodeRegion),
			PkName: "id",
		},
	}).(*NodeRegionDAO)
}

var SharedNodeRegionDAO *NodeRegionDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeRegionDAO = NewNodeRegionDAO()
	})
}

// 启用条目
func (this *NodeRegionDAO) EnableNodeRegion(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeRegionStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *NodeRegionDAO) DisableNodeRegion(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeRegionStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *NodeRegionDAO) FindEnabledNodeRegion(tx *dbs.Tx, id int64) (*NodeRegion, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeRegionStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeRegion), err
}

// 根据主键查找名称
func (this *NodeRegionDAO) FindNodeRegionName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建区域
func (this *NodeRegionDAO) CreateRegion(tx *dbs.Tx, adminId int64, name string, description string) (int64, error) {
	op := NewNodeRegionOperator()
	op.AdminId = adminId
	op.Name = name
	op.Description = description
	op.State = NodeRegionStateEnabled
	op.IsOn = true
	return this.SaveInt64(tx, op)
}

// 修改区域
func (this *NodeRegionDAO) UpdateRegion(tx *dbs.Tx, regionId int64, name string, description string, isOn bool) error {
	if regionId <= 0 {
		return errors.New("invalid regionId")
	}
	op := NewNodeRegionOperator()
	op.Id = regionId
	op.Name = name
	op.Description = description
	op.IsOn = isOn
	return this.Save(tx, op)
}

// 列出所有区域
func (this *NodeRegionDAO) FindAllEnabledRegions(tx *dbs.Tx) (result []*NodeRegion, err error) {
	_, err = this.Query(tx).
		State(NodeRegionStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 列出所有价格
func (this *NodeRegionDAO) FindAllEnabledRegionPrices(tx *dbs.Tx) (result []*NodeRegion, err error) {
	_, err = this.Query(tx).
		State(NodeRegionStateEnabled).
		Desc("order").
		AscPk().
		Result("id", "prices").
		Slice(&result).
		FindAll()
	return
}

// 列出所有启用的区域
func (this *NodeRegionDAO) FindAllEnabledAndOnRegions(tx *dbs.Tx) (result []*NodeRegion, err error) {
	_, err = this.Query(tx).
		State(NodeRegionStateEnabled).
		Attr("isOn", true).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 排序
func (this *NodeRegionDAO) UpdateRegionOrders(tx *dbs.Tx, regionIds []int64) error {
	order := len(regionIds)
	for _, regionId := range regionIds {
		_, err := this.Query(tx).
			Pk(regionId).
			Set("order", order).
			Update()
		if err != nil {
			return err
		}
		order--
	}
	return nil
}

// 修改价格项价格
func (this *NodeRegionDAO) UpdateRegionItemPrice(tx *dbs.Tx, regionId int64, itemId int64, price float32) error {
	one, err := this.Query(tx).
		Pk(regionId).
		Result("prices").
		Find()
	if err != nil {
		return err
	}
	if one == nil {
		return nil
	}
	prices := one.(*NodeRegion).Prices
	pricesMap := map[string]float32{}
	if len(prices) > 0 && prices != "null" {
		err = json.Unmarshal([]byte(prices), &pricesMap)
		if err != nil {
			return err
		}
	}
	pricesMap[numberutils.FormatInt64(itemId)] = price
	pricesJSON, err := json.Marshal(pricesMap)
	if err != nil {
		return err
	}
	_, err = this.Query(tx).
		Pk(regionId).
		Set("prices", pricesJSON).
		Update()
	return err
}
