package models

//
type File struct {
	Id          uint32 `field:"id"`          // ID
	Description string `field:"description"` // 文件描述
	Filename    string `field:"filename"`    // 文件名
	Size        uint32 `field:"size"`        // 文件尺寸
	CreatedAt   uint32 `field:"createdAt"`   // 创建时间
	Order       uint32 `field:"order"`       // 排序
	Type        string `field:"type"`        // 类型
	State       uint8  `field:"state"`       // 状态
}

type FileOperator struct {
	Id          interface{} // ID
	Description interface{} // 文件描述
	Filename    interface{} // 文件名
	Size        interface{} // 文件尺寸
	CreatedAt   interface{} // 创建时间
	Order       interface{} // 排序
	Type        interface{} // 类型
	State       interface{} // 状态
}

func NewFileOperator() *FileOperator {
	return &FileOperator{}
}
