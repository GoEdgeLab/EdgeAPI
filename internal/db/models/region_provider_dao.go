package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	RegionProviderStateEnabled  = 1 // 已启用
	RegionProviderStateDisabled = 0 // 已禁用
)

type RegionProviderDAO dbs.DAO

func NewRegionProviderDAO() *RegionProviderDAO {
	return dbs.NewDAO(&RegionProviderDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionProviders",
			Model:  new(RegionProvider),
			PkName: "id",
		},
	}).(*RegionProviderDAO)
}

var SharedRegionProviderDAO *RegionProviderDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionProviderDAO = NewRegionProviderDAO()
	})
}

// 启用条目
func (this *RegionProviderDAO) EnableRegionProvider(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionProviderStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *RegionProviderDAO) DisableRegionProvider(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionProviderStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *RegionProviderDAO) FindEnabledRegionProvider(id uint32) (*RegionProvider, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", RegionProviderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionProvider), err
}

// 根据主键查找名称
func (this *RegionProviderDAO) FindRegionProviderName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
