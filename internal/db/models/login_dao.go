package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

const (
	LoginStateEnabled  = 1 // 已启用
	LoginStateDisabled = 0 // 已禁用
)

type LoginType = string

const (
	LoginTypeOTP LoginType = "otp"
)

type LoginDAO dbs.DAO

func NewLoginDAO() *LoginDAO {
	return dbs.NewDAO(&LoginDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeLogins",
			Model:  new(Login),
			PkName: "id",
		},
	}).(*LoginDAO)
}

var SharedLoginDAO *LoginDAO

func init() {
	dbs.OnReady(func() {
		SharedLoginDAO = NewLoginDAO()
	})
}

// 启用条目
func (this *LoginDAO) EnableLogin(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", LoginStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *LoginDAO) DisableLogin(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", LoginStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *LoginDAO) FindEnabledLogin(id int64) (*Login, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", LoginStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Login), err
}

// 创建认证
func (this *LoginDAO) CreateLogin(Id int64, loginType LoginType, params maps.Map) (int64, error) {
	if Id <= 0 {
		return 0, errors.New("invalid Id")
	}
	if params == nil {
		params = maps.Map{}
	}
	op := NewLoginOperator()
	op.Id = Id
	op.Type = loginType
	op.Params = params.AsJSON()
	op.State = LoginStateEnabled
	op.IsOn = true
	return this.SaveInt64(op)
}

// 修改认证
func (this *LoginDAO) UpdateLogin(adminId int64, loginType LoginType, params maps.Map, isOn bool) error {
	// 是否已经存在
	loginId, err := this.Query().
		Attr("adminId", adminId).
		Attr("type", loginType).
		State(LoginStateEnabled).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return err
	}
	op := NewLoginOperator()
	if loginId > 0 {
		op.Id = loginId
	} else {
		op.AdminId = adminId
		op.Type = loginType
		op.State = LoginStateEnabled
	}

	if params == nil {
		params = maps.Map{}
	}

	op.IsOn = isOn
	op.Params = params.AsJSON()
	return this.Save(op)
}

// 禁用相关认证
func (this *LoginDAO) DisableLoginWithAdminId(adminId int64, loginType LoginType) error {
	_, err := this.Query().
		Attr("adminId", adminId).
		Attr("type", loginType).
		Set("isOn", false).
		Update()
	return err
}

// 查找管理员相关的认证
func (this *LoginDAO) FindEnabledLoginWithAdminId(adminId int64, loginType LoginType) (*Login, error) {
	one, err := this.Query().
		Attr("adminId", adminId).
		Attr("type", loginType).
		State(LoginStateEnabled).
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*Login), nil
}

// 检查某个认证是否启用
func (this *LoginDAO) CheckLoginIsOn(adminId int64, loginType LoginType) (bool, error) {
	return this.Query().
		Attr("adminId", adminId).
		Attr("type", loginType).
		State(LoginStateEnabled).
		Attr("isOn", true).
		Exist()
}
