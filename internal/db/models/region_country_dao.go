package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	RegionCountryStateEnabled  = 1 // 已启用
	RegionCountryStateDisabled = 0 // 已禁用
)

type RegionCountryDAO dbs.DAO

func NewRegionCountryDAO() *RegionCountryDAO {
	return dbs.NewDAO(&RegionCountryDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionCountries",
			Model:  new(RegionCountry),
			PkName: "id",
		},
	}).(*RegionCountryDAO)
}

var SharedRegionCountryDAO *RegionCountryDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionCountryDAO = NewRegionCountryDAO()
	})
}

// 启用条目
func (this *RegionCountryDAO) EnableRegionCountry(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionCountryStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *RegionCountryDAO) DisableRegionCountry(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionCountryStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *RegionCountryDAO) FindEnabledRegionCountry(id uint32) (*RegionCountry, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", RegionCountryStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionCountry), err
}

// 根据主键查找名称
func (this *RegionCountryDAO) FindRegionCountryName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
