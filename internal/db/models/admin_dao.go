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
func (this *AdminDAO) EnableAdmin(id int64) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", AdminStateEnabled).
		Update()
}

// 禁用条目
func (this *AdminDAO) DisableAdmin(id int64) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", AdminStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *AdminDAO) FindEnabledAdmin(id int64) (*Admin, error) {
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
func (this *AdminDAO) ExistEnabledAdmin(adminId int64) (bool, error) {
	return this.Query().
		Pk(adminId).
		State(AdminStateEnabled).
		Exist()
}

// 获取管理员名称
func (this *AdminDAO) FindAdminFullname(adminId int64) (string, error) {
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
		Attr("isOn", true).
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
	err := this.Save(op)
	return err
}

// 创建管理员
func (this *AdminDAO) CreateAdmin(username string, password string, fullname string, isSuper bool, modulesJSON []byte) (int64, error) {
	op := NewAdminOperator()
	op.IsOn = true
	op.State = AdminStateEnabled
	op.Username = username
	op.Password = stringutil.Md5(password)
	op.Fullname = fullname
	op.IsSuper = isSuper
	if len(modulesJSON) > 0 {
		op.Modules = modulesJSON
	} else {
		op.Modules = "[]"
	}
	err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改管理员个人资料
func (this *AdminDAO) UpdateAdminInfo(adminId int64, fullname string) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	op := NewAdminOperator()
	op.Id = adminId
	op.Fullname = fullname
	err := this.Save(op)
	return err
}

// 修改管理员详细信息
func (this *AdminDAO) UpdateAdmin(adminId int64, username string, password string, fullname string, isSuper bool, modulesJSON []byte, isOn bool) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	op := NewAdminOperator()
	op.Id = adminId
	op.Fullname = fullname
	op.Username = username
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
	err := this.Save(op)
	return err
}

// 检查用户名是否存在
func (this *AdminDAO) CheckAdminUsername(adminId int64, username string) (bool, error) {
	query := this.Query().
		State(AdminStateEnabled).
		Attr("username", username)
	if adminId > 0 {
		query.
			Where("id!=:id").
			Param("id", adminId)
	}
	return query.Exist()
}

// 修改管理员登录信息
func (this *AdminDAO) UpdateAdminLogin(adminId int64, username string, password string) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	op := NewAdminOperator()
	op.Id = adminId
	op.Username = username
	if len(password) > 0 {
		op.Password = stringutil.Md5(password)
	}
	err := this.Save(op)
	return err
}

// 修改管理员可以管理的模块
func (this *AdminDAO) UpdateAdminModules(adminId int64, allowModulesJSON []byte) error {
	if adminId <= 0 {
		return errors.New("invalid adminId")
	}
	op := NewAdminOperator()
	op.Id = adminId
	op.Modules = allowModulesJSON
	err := this.Save(op)
	if err != nil {
		return err
	}
	return nil
}

// 查询所有管理的权限
func (this *AdminDAO) FindAllAdminModules() (result []*Admin, err error) {
	_, err = this.Query().
		State(AdminStateEnabled).
		Attr("isOn", true).
		Result("id", "modules", "isSuper").
		Slice(&result).
		FindAll()
	return
}

// 计算所有管理员数量
func (this *AdminDAO) CountAllEnabledAdmins() (int64, error) {
	return this.Query().
		State(AdminStateEnabled).
		Count()
}

// 列出单页的管理员
func (this *AdminDAO) ListEnabledAdmins(offset int64, size int64) (result []*Admin, err error) {
	_, err = this.Query().
		State(AdminStateEnabled).
		Result("id", "isOn", "username", "fullname", "isSuper", "createdAt").
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
