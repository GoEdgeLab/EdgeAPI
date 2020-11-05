package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
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
