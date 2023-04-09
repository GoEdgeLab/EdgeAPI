package models

import (
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
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

// EnableAdmin 启用条目
func (this *AdminDAO) EnableAdmin(tx *dbs.Tx, id int64) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", AdminStateEnabled).
		Update()
}

// DisableAdmin 禁用条目
func (this *AdminDAO) DisableAdmin(tx *dbs.Tx, adminId int64) error {
	err := this.Query(tx).
		Pk(adminId).
		Set("state", AdminStateDisabled).
		UpdateQuickly()
	if err != nil {
		return err
	}

	// 删除AccessTokens
	return SharedAPIAccessTokenDAO.DeleteAccessTokens(tx, adminId, 0)
}

// FindEnabledAdmin 查找启用中的条目
func (this *AdminDAO) FindEnabledAdmin(tx *dbs.Tx, id int64) (*Admin, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", AdminStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Admin), err
}

// FindBasicAdmin 查找管理员基本信息
func (this *AdminDAO) FindBasicAdmin(tx *dbs.Tx, id int64) (*Admin, error) {
	result, err := this.Query(tx).
		Result("id", "username", "fullname").
		Pk(id).
		Attr("state", AdminStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Admin), err
}

// ExistEnabledAdmin 检查管理员是否存在
func (this *AdminDAO) ExistEnabledAdmin(tx *dbs.Tx, adminId int64) (bool, error) {
	return this.Query(tx).
		Pk(adminId).
		State(AdminStateEnabled).
		Exist()
}

// FindAdminFullname 获取管理员名称
func (this *AdminDAO) FindAdminFullname(tx *dbs.Tx, adminId int64) (string, error) {
	return this.Query(tx).
		Pk(adminId).
		Result("fullname").
		FindStringCol("")
}

// CheckAdminPassword 检查用户名、密码
func (this *AdminDAO) CheckAdminPassword(tx *dbs.Tx, username string, encryptedPassword string) (int64, error) {
	if len(username) == 0 || len(encryptedPassword) == 0 {
		return 0, nil
	}
	return this.Query(tx).
		Attr("username", username).
		Attr("password", encryptedPassword).
		Attr("state", AdminStateEnabled).
		Attr("isOn", true).
		Attr("canLogin", 1).
		ResultPk().
		FindInt64Col(0)
}

// FindAdminIdWithUsername 根据用户名查询管理员ID
func (this *AdminDAO) FindAdminIdWithUsername(tx *dbs.Tx, username string) (int64, error) {
	one, err := this.Query(tx).
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

// UpdateAdminPassword 更改管理员密码
func (this *AdminDAO) UpdateAdminPassword(tx *dbs.Tx, adminId int64, password string) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	var op = NewAdminOperator()
	op.Id = adminId
	op.Password = stringutil.Md5(password)
	err := this.Save(tx, op)
	return err
}

// CreateAdmin 创建管理员
func (this *AdminDAO) CreateAdmin(tx *dbs.Tx, username string, canLogin bool, password string, fullname string, isSuper bool, modulesJSON []byte) (int64, error) {
	var op = NewAdminOperator()
	op.IsOn = true
	op.State = AdminStateEnabled
	op.Username = username
	op.CanLogin = canLogin
	op.Password = stringutil.Md5(password)
	op.Fullname = fullname
	op.IsSuper = isSuper
	if len(modulesJSON) > 0 {
		op.Modules = modulesJSON
	} else {
		op.Modules = "[]"
	}
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateAdminInfo 修改管理员个人资料
func (this *AdminDAO) UpdateAdminInfo(tx *dbs.Tx, adminId int64, fullname string) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	var op = NewAdminOperator()
	op.Id = adminId
	op.Fullname = fullname
	err := this.Save(tx, op)
	return err
}

// UpdateAdmin 修改管理员详细信息
func (this *AdminDAO) UpdateAdmin(tx *dbs.Tx, adminId int64, username string, canLogin bool, password string, fullname string, isSuper bool, modulesJSON []byte, isOn bool) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	var op = NewAdminOperator()
	op.Id = adminId
	op.Fullname = fullname
	op.Username = username
	op.CanLogin = canLogin
	if len(password) > 0 {
		op.Password = stringutil.Md5(password)
	}
	op.IsSuper = isSuper
	if len(modulesJSON) > 0 {
		op.Modules = modulesJSON
	} else {
		op.Modules = "[]"
	}
	op.IsOn = isOn
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	if !isOn {
		// 删除AccessTokens
		err = SharedAPIAccessTokenDAO.DeleteAccessTokens(tx, adminId, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckAdminUsername 检查用户名是否存在
func (this *AdminDAO) CheckAdminUsername(tx *dbs.Tx, adminId int64, username string) (bool, error) {
	query := this.Query(tx).
		State(AdminStateEnabled).
		Attr("username", username)
	if adminId > 0 {
		query.
			Where("id!=:id").
			Param("id", adminId)
	}
	return query.Exist()
}

// UpdateAdminLogin 修改管理员登录信息
func (this *AdminDAO) UpdateAdminLogin(tx *dbs.Tx, adminId int64, username string, password string) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	var op = NewAdminOperator()
	op.Id = adminId
	op.Username = username
	if len(password) > 0 {
		op.Password = stringutil.Md5(password)
	}
	err := this.Save(tx, op)
	return err
}

// UpdateAdminModules 修改管理员可以管理的模块
func (this *AdminDAO) UpdateAdminModules(tx *dbs.Tx, adminId int64, allowModulesJSON []byte) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	var op = NewAdminOperator()
	op.Id = adminId
	op.Modules = allowModulesJSON
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return nil
}

// FindAllAdminModules 查询所有管理的权限
func (this *AdminDAO) FindAllAdminModules(tx *dbs.Tx) (result []*Admin, err error) {
	_, err = this.Query(tx).
		State(AdminStateEnabled).
		Attr("isOn", true).
		Result("id", "modules", "isSuper", "fullname", "theme").
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledAdmins 计算所有管理员数量
func (this *AdminDAO) CountAllEnabledAdmins(tx *dbs.Tx, keyword string, hasWeakPasswords bool) (int64, error) {
	var query = this.Query(tx)
	if len(keyword) > 0 {
		query.Where("(username LIKE :keyword OR fullname LIKE :keyword)")
		query.Param("keyword", dbutils.QuoteLike(keyword))
	}
	if hasWeakPasswords {
		query.Attr("password", weakPasswords)
		query.Attr("isOn", true)
	}
	return query.
		State(AdminStateEnabled).
		Count()
}

// ListEnabledAdmins 列出单页的管理员
func (this *AdminDAO) ListEnabledAdmins(tx *dbs.Tx, keyword string, hasWeakPasswords bool, offset int64, size int64) (result []*Admin, err error) {
	var query = this.Query(tx)
	if len(keyword) > 0 {
		query.Where("(username LIKE :keyword OR fullname LIKE :keyword)")
		query.Param("keyword", dbutils.QuoteLike(keyword))
	}
	if hasWeakPasswords {
		query.Attr("password", weakPasswords)
		query.Attr("isOn", true)
	}

	_, err = query.
		State(AdminStateEnabled).
		Result("id", "isOn", "username", "fullname", "isSuper", "createdAt", "canLogin", "password").
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// UpdateAdminTheme 设置管理员Theme
func (this *AdminDAO) UpdateAdminTheme(tx *dbs.Tx, adminId int64, theme string) error {
	return this.Query(tx).
		Pk(adminId).
		Set("theme", theme).
		UpdateQuickly()
}

// CheckSuperAdmin 检查管理员是否为超级管理员
func (this *AdminDAO) CheckSuperAdmin(tx *dbs.Tx, adminId int64) (bool, error) {
	if adminId <= 0 {
		return false, nil
	}
	return this.Query(tx).
		Pk(adminId).
		State(AdminStateEnabled).
		Attr("isSuper", true).
		Exist()
}
