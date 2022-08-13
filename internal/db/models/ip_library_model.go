package models

// IPLibrary IP库
type IPLibrary struct {
	Id        uint32 `field:"id"`        // ID
	AdminId   uint32 `field:"adminId"`   // 管理员ID
	FileId    uint32 `field:"fileId"`    // 文件ID
	Type      string `field:"type"`      // 类型
	Name      string `field:"name"`      // 名称
	IsPublic  bool   `field:"isPublic"`  // 是否公用
	State     uint8  `field:"state"`     // 状态
	CreatedAt uint64 `field:"createdAt"` // 创建时间
}

type IPLibraryOperator struct {
	Id        any // ID
	AdminId   any // 管理员ID
	FileId    any // 文件ID
	Type      any // 类型
	Name      any // 名称
	IsPublic  any // 是否公用
	State     any // 状态
	CreatedAt any // 创建时间
}

func NewIPLibraryOperator() *IPLibraryOperator {
	return &IPLibraryOperator{}
}
