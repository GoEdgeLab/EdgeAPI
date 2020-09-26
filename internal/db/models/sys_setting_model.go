package models

// 系统配置
type SysSetting struct {
	Id     uint32 `field:"id"`     // ID
	UserId uint32 `field:"userId"` // 用户ID
	Code   string `field:"code"`   // 代号
	Value  string `field:"value"`  // 配置值
}

type SysSettingOperator struct {
	Id     interface{} // ID
	UserId interface{} // 用户ID
	Code   interface{} // 代号
	Value  interface{} // 配置值
}

func NewSysSettingOperator() *SysSettingOperator {
	return &SysSettingOperator{}
}
