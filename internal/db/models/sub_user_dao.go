package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	SubUserStateEnabled  = 1 // 已启用
	SubUserStateDisabled = 0 // 已禁用
)

type SubUserDAO dbs.DAO

func NewSubUserDAO() *SubUserDAO {
	return dbs.NewDAO(&SubUserDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeSubUsers",
			Model:  new(SubUser),
			PkName: "id",
		},
	}).(*SubUserDAO)
}

var SharedSubUserDAO *SubUserDAO

func init() {
	dbs.OnReady(func() {
		SharedSubUserDAO = NewSubUserDAO()
	})
}

// 启用条目
func (this *SubUserDAO) EnableSubUser(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", SubUserStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *SubUserDAO) DisableSubUser(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", SubUserStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *SubUserDAO) FindEnabledSubUser(tx *dbs.Tx, id uint32) (*SubUser, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", SubUserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*SubUser), err
}

// 根据主键查找名称
func (this *SubUserDAO) FindSubUserName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}
