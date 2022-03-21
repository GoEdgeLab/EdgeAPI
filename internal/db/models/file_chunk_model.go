package models

//
type FileChunk struct {
	Id     uint32 `field:"id"`     // ID
	FileId uint32 `field:"fileId"` // 文件ID
	Data   []byte `field:"data"`   // 分块内容
}

type FileChunkOperator struct {
	Id     interface{} // ID
	FileId interface{} // 文件ID
	Data   interface{} // 分块内容
}

func NewFileChunkOperator() *FileChunkOperator {
	return &FileChunkOperator{}
}
