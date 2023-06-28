package models

import "github.com/iwind/TeaGo/dbs"

const (
	AdminFieldId        dbs.FieldName = "id"        // ID
	AdminFieldIsOn      dbs.FieldName = "isOn"      // 是否启用
	AdminFieldUsername  dbs.FieldName = "username"  // 用户名
	AdminFieldPassword  dbs.FieldName = "password"  // 密码
	AdminFieldFullname  dbs.FieldName = "fullname"  // 全名
	AdminFieldIsSuper   dbs.FieldName = "isSuper"   // 是否为超级管理员
	AdminFieldCreatedAt dbs.FieldName = "createdAt" // 创建时间
	AdminFieldUpdatedAt dbs.FieldName = "updatedAt" // 修改时间
	AdminFieldState     dbs.FieldName = "state"     // 状态
	AdminFieldModules   dbs.FieldName = "modules"   // 允许的模块
	AdminFieldCanLogin  dbs.FieldName = "canLogin"  // 是否可以登录
	AdminFieldTheme     dbs.FieldName = "theme"     // 模板设置
	AdminFieldLang      dbs.FieldName = "lang"      // 语言代号
)

// Admin 管理员
type Admin struct {
	Id        uint32   `field:"id"`        // ID
	IsOn      bool     `field:"isOn"`      // 是否启用
	Username  string   `field:"username"`  // 用户名
	Password  string   `field:"password"`  // 密码
	Fullname  string   `field:"fullname"`  // 全名
	IsSuper   bool     `field:"isSuper"`   // 是否为超级管理员
	CreatedAt uint64   `field:"createdAt"` // 创建时间
	UpdatedAt uint64   `field:"updatedAt"` // 修改时间
	State     uint8    `field:"state"`     // 状态
	Modules   dbs.JSON `field:"modules"`   // 允许的模块
	CanLogin  bool     `field:"canLogin"`  // 是否可以登录
	Theme     string   `field:"theme"`     // 模板设置
	Lang      string   `field:"lang"`      // 语言代号
}

type AdminOperator struct {
	Id        any // ID
	IsOn      any // 是否启用
	Username  any // 用户名
	Password  any // 密码
	Fullname  any // 全名
	IsSuper   any // 是否为超级管理员
	CreatedAt any // 创建时间
	UpdatedAt any // 修改时间
	State     any // 状态
	Modules   any // 允许的模块
	CanLogin  any // 是否可以登录
	Theme     any // 模板设置
	Lang      any // 语言代号
}

func NewAdminOperator() *AdminOperator {
	return &AdminOperator{}
}
