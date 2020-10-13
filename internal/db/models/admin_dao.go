package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
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

var SharedAdminDAO *AdminDAO

func init() {
	dbs.OnReady(func() {
		SharedAdminDAO = NewAdminDAO()
	})
}

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
func (this *AdminDAO) CheckAdminPassword(username string, encryptedPassword string) (int64, error) {
	if len(username) == 0 || len(encryptedPassword) == 0 {
		return 0, nil
	}
	return this.Query().
		Attr("username", username).
		Attr("password", encryptedPassword).
		Attr("state", AdminStateEnabled).
		ResultPk().
		FindInt64Col(0)
}

// 根据用户名查询管理员ID
func (this *AdminDAO) FindAdminIdWithUsername(username string) (int64, error) {
	one, err := this.Query().
		Attr("username", username).
		State(AdminStateEnabled).
		ResultPk().
		Find()
	if err != nil {
		return 0, err
	}
	if one == nil {
		return 0, nil
	}
	return int64(one.(*Admin).Id), nil
}

// 更改管理员密码
func (this *AdminDAO) UpdateAdminPassword(adminId int64, password string) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	op := NewAdminOperator()
	op.Id = adminId
	op.Password = stringutil.Md5(password)
	_, err := this.Save(op)
	return err
}

// 创建管理员
func (this *AdminDAO) CreateAdmin(username string, password string, fullname string) (int64, error) {
	op := NewAdminOperator()
	op.Username = username
	op.Password = stringutil.Md5(password)
	op.Fullname = fullname
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}
