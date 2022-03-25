package models

// ScriptHistory 脚本历史记录
type ScriptHistory struct {
	Id       uint32 `field:"id"`       // ID
	UserId   uint64 `field:"userId"`   // 用户ID
	ScriptId uint64 `field:"scriptId"` // 脚本ID
	Filename string `field:"filename"` // 文件名
	Code     string `field:"code"`     // 代码
	Version  uint64 `field:"version"`  // 版本号
}

type ScriptHistoryOperator struct {
	Id       interface{} // ID
	UserId   interface{} // 用户ID
	ScriptId interface{} // 脚本ID
	Filename interface{} // 文件名
	Code     interface{} // 代码
	Version  interface{} // 版本号
}

func NewScriptHistoryOperator() *ScriptHistoryOperator {
	return &ScriptHistoryOperator{}
}
