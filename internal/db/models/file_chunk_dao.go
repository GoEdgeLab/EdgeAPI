package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type FileChunkDAO dbs.DAO

func NewFileChunkDAO() *FileChunkDAO {
	return dbs.NewDAO(&FileChunkDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeFileChunks",
			Model:  new(FileChunk),
			PkName: "id",
		},
	}).(*FileChunkDAO)
}

var SharedFileChunkDAO = NewFileChunkDAO()

// 创建文件Chunk
func (this *FileChunkDAO) CreateFileChunk(fileId int, data []byte) error {
	op := NewFileChunkOperator()
	op.FileId = fileId
	op.Data = data
	_, err := this.Save(op)
	return err
}

// 列出所有的文件Chunk
func (this *FileChunkDAO) FindAllFileChunks(fileId int) (result []*FileChunk, err error) {
	_, err = this.Query().
		Attr("fileId", fileId).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 删除以前的文件
func (this *FileChunkDAO) DeleteFileChunks(fileId int) error {
	if fileId <= 0 {
		return errors.New("invalid fileId")
	}
	_, err := this.Query().
		Attr("fileId", fileId).
		Delete()
	return err
}
