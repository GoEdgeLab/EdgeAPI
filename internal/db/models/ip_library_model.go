package models

// IP库
type IPLibrary struct {
	Id        uint32 `field:"id"`        // ID
	AdminId   uint32 `field:"adminId"`   // 管理员ID
	FileId    uint32 `field:"fileId"`    // 文件ID
	Type      string `field:"type"`      // 类型
	State     uint8  `field:"state"`     // 状态
	CreatedAt uint64 `field:"createdAt"` // 创建时间
}

type IPLibraryOperator struct {
	Id        interface{} // ID
	AdminId   interface{} // 管理员ID
	FileId    interface{} // 文件ID
	Type      interface{} // 类型
	State     interface{} // 状态
	CreatedAt interface{} // 创建时间
}

func NewIPLibraryOperator() *IPLibraryOperator {
	return &IPLibraryOperator{}
}
