package models

// Script 脚本库
type Script struct {
	Id        uint64 `field:"id"`        // ID
	UserId    uint64 `field:"userId"`    // 用户ID
	IsOn      bool   `field:"isOn"`      // 是否启用
	Name      string `field:"name"`      // 名称
	Filename  string `field:"filename"`  // 文件名
	Code      string `field:"code"`      // 代码
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	UpdatedAt uint64 `field:"updatedAt"` // 修改时间
	State     uint8  `field:"state"`     // 是否启用
}

type ScriptOperator struct {
	Id        interface{} // ID
	UserId    interface{} // 用户ID
	IsOn      interface{} // 是否启用
	Name      interface{} // 名称
	Filename  interface{} // 文件名
	Code      interface{} // 代码
	CreatedAt interface{} // 创建时间
	UpdatedAt interface{} // 修改时间
	State     interface{} // 是否启用
}

func NewScriptOperator() *ScriptOperator {
	return &ScriptOperator{}
}
