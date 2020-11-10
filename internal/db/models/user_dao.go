package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	UserStateEnabled  = 1 // 已启用
	UserStateDisabled = 0 // 已禁用
)

type UserDAO dbs.DAO

func NewUserDAO() *UserDAO {
	return dbs.NewDAO(&UserDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUsers",
			Model:  new(User),
			PkName: "id",
		},
	}).(*UserDAO)
}

var SharedUserDAO *UserDAO

func init() {
	dbs.OnReady(func() {
		SharedUserDAO = NewUserDAO()
	})
}

// 启用条目
func (this *UserDAO) EnableUser(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", UserStateEnabled).
		Update()
}

// 禁用条目
func (this *UserDAO) DisableUser(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", UserStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *UserDAO) FindEnabledUser(id uint32) (*User, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", UserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*User), err
}

// 获取管理员名称
func (this *UserDAO) FindUserFullname(userId int64) (string, error) {
	return this.Query().
		Pk(userId).
		Result("fullname").
		FindStringCol("")
}
