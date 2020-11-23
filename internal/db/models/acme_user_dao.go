package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ACMEUserStateEnabled  = 1 // 已启用
	ACMEUserStateDisabled = 0 // 已禁用
)

type ACMEUserDAO dbs.DAO

func NewACMEUserDAO() *ACMEUserDAO {
	return dbs.NewDAO(&ACMEUserDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeACMEUsers",
			Model:  new(ACMEUser),
			PkName: "id",
		},
	}).(*ACMEUserDAO)
}

var SharedACMEUserDAO *ACMEUserDAO

func init() {
	dbs.OnReady(func() {
		SharedACMEUserDAO = NewACMEUserDAO()
	})
}

// 启用条目
func (this *ACMEUserDAO) EnableACMEUser(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ACMEUserStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *ACMEUserDAO) DisableACMEUser(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ACMEUserStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *ACMEUserDAO) FindEnabledACMEUser(id int64) (*ACMEUser, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ACMEUserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ACMEUser), err
}
