package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ProviderStateEnabled  = 1 // 已启用
	ProviderStateDisabled = 0 // 已禁用
)

type ProviderDAO dbs.DAO

func NewProviderDAO() *ProviderDAO {
	return dbs.NewDAO(&ProviderDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeProviders",
			Model:  new(Provider),
			PkName: "id",
		},
	}).(*ProviderDAO)
}

var SharedProviderDAO = NewProviderDAO()

// 启用条目
func (this *ProviderDAO) EnableProvider(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ProviderStateEnabled).
		Update()
}

// 禁用条目
func (this *ProviderDAO) DisableProvider(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ProviderStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *ProviderDAO) FindEnabledProvider(id uint32) (*Provider, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ProviderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Provider), err
}
