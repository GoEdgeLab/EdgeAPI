package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	RegionProvinceStateEnabled  = 1 // 已启用
	RegionProvinceStateDisabled = 0 // 已禁用
)

type RegionProvinceDAO dbs.DAO

func NewRegionProvinceDAO() *RegionProvinceDAO {
	return dbs.NewDAO(&RegionProvinceDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionProvinces",
			Model:  new(RegionProvince),
			PkName: "id",
		},
	}).(*RegionProvinceDAO)
}

var SharedRegionProvinceDAO *RegionProvinceDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionProvinceDAO = NewRegionProvinceDAO()
	})
}

// 启用条目
func (this *RegionProvinceDAO) EnableRegionProvince(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionProvinceStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *RegionProvinceDAO) DisableRegionProvince(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionProvinceStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *RegionProvinceDAO) FindEnabledRegionProvince(id uint32) (*RegionProvince, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", RegionProvinceStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionProvince), err
}

// 根据主键查找名称
func (this *RegionProvinceDAO) FindRegionProvinceName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
