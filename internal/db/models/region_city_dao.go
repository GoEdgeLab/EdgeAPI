package models

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	RegionCityStateEnabled  = 1 // 已启用
	RegionCityStateDisabled = 0 // 已禁用
)

type RegionCityDAO dbs.DAO

func NewRegionCityDAO() *RegionCityDAO {
	return dbs.NewDAO(&RegionCityDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionCities",
			Model:  new(RegionCity),
			PkName: "id",
		},
	}).(*RegionCityDAO)
}

var SharedRegionCityDAO *RegionCityDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionCityDAO = NewRegionCityDAO()
	})
}

// 启用条目
func (this *RegionCityDAO) EnableRegionCity(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionCityStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *RegionCityDAO) DisableRegionCity(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionCityStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *RegionCityDAO) FindEnabledRegionCity(id uint32) (*RegionCity, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", RegionCityStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionCity), err
}

// 根据主键查找名称
func (this *RegionCityDAO) FindRegionCityName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 根据数据ID查找城市
func (this *RegionCityDAO) FindCityWithDataId(dataId string) (int64, error) {
	return this.Query().
		Attr("dataId", dataId).
		ResultPk().
		FindInt64Col(0)
}

// 创建城市
func (this *RegionCityDAO) CreateCity(provinceId int64, name string, dataId string) (int64, error) {
	op := NewRegionCityOperator()
	op.ProvinceId = provinceId
	op.Name = name
	op.DataId = dataId
	op.State = RegionCityStateEnabled

	codes := []string{name}
	codesJSON, err := json.Marshal(codes)
	if err != nil {
		return 0, err
	}
	op.Codes = codesJSON
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}
