package models

// 文件管理
type File struct {
	Id          uint32 `field:"id"`          // ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	UserId      uint32 `field:"userId"`      // 用户ID
	Description string `field:"description"` // 文件描述
	Filename    string `field:"filename"`    // 文件名
	Size        uint32 `field:"size"`        // 文件尺寸
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	Order       uint32 `field:"order"`       // 排序
	Type        string `field:"type"`        // 类型
	State       uint8  `field:"state"`       // 状态
	IsFinished  bool   `field:"isFinished"`  // 是否已完成上传
	IsPublic    bool   `field:"isPublic"`    // 是否可以公开访问
}

type FileOperator struct {
	Id          interface{} // ID
	AdminId     interface{} // 管理员ID
	UserId      interface{} // 用户ID
	Description interface{} // 文件描述
	Filename    interface{} // 文件名
	Size        interface{} // 文件尺寸
	CreatedAt   interface{} // 创建时间
	Order       interface{} // 排序
	Type        interface{} // 类型
	State       interface{} // 状态
	IsFinished  interface{} // 是否已完成上传
	IsPublic    interface{} // 是否可以公开访问
}

func NewFileOperator() *FileOperator {
	return &FileOperator{}
}
