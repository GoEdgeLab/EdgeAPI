package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	AdminStateEnabled  = 1 // 已启用
	AdminStateDisabled = 0 // 已禁用
)

type AdminDAO dbs.DAO

func NewAdminDAO() *AdminDAO {
	return dbs.NewDAO(&AdminDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeAdmins",
			Model:  new(Admin),
			PkName: "id",
		},
	}).(*AdminDAO)
}

var SharedAdminDAO = NewAdminDAO()

// 启用条目
func (this *AdminDAO) EnableAdmin(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", AdminStateEnabled).
		Update()
}

// 禁用条目
func (this *AdminDAO) DisableAdmin(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", AdminStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *AdminDAO) FindEnabledAdmin(id uint32) (*Admin, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", AdminStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Admin), err
}

// 检查管理员是否存在
func (this *AdminDAO) ExistEnabledAdmin(adminId int) (bool, error) {
	return this.Query().
		Pk(adminId).
		State(AdminStateEnabled).
		Exist()
}

// 获取管理员名称
func (this *AdminDAO) FindAdminFullname(adminId int) (string, error) {
	return this.Query().
		Pk(adminId).
		Result("fullname").
		FindStringCol("")
}

// 检查用户名、密码
func (this *AdminDAO) CheckAdminPassword(username string, encryptedPassword string) (int, error) {
	if len(username) == 0 || len(encryptedPassword) == 0 {
		return 0, nil
	}
	return this.Query().
		Attr("username", username).
		Attr("password", encryptedPassword).
		Attr("state", AdminStateEnabled).
		ResultPk().
		FindIntCol(0)
}
