package models

import "github.com/iwind/TeaGo/dbs"

// SysSetting 系统配置
type SysSetting struct {
	Id     uint32   `field:"id"`     // ID
	UserId uint32   `field:"userId"` // 用户ID
	Code   string   `field:"code"`   // 代号
	Value  dbs.JSON `field:"value"`  // 配置值
}

type SysSettingOperator struct {
	Id     any // ID
	UserId any // 用户ID
	Code   any // 代号
	Value  any // 配置值
}

func NewSysSettingOperator() *SysSettingOperator {
	return &SysSettingOperator{}
}
