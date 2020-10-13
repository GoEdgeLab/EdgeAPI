package models

// 数据库结构版本
type Version struct {
	Id      uint64 `field:"id"`      // ID
	Version string `field:"version"` //
}

type VersionOperator struct {
	Id      interface{} // ID
	Version interface{} //
}

func NewVersionOperator() *VersionOperator {
	return &VersionOperator{}
}
