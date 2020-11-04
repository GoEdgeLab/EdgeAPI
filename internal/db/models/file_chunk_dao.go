package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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

var SharedFileChunkDAO *FileChunkDAO

func init() {
	dbs.OnReady(func() {
		SharedFileChunkDAO = NewFileChunkDAO()
	})
}

// 创建文件Chunk
func (this *FileChunkDAO) CreateFileChunk(fileId int64, data []byte) (int64, error) {
	op := NewFileChunkOperator()
	op.FileId = fileId
	op.Data = data
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 列出所有的文件Chunk
func (this *FileChunkDAO) FindAllFileChunks(fileId int64) (result []*FileChunk, err error) {
	_, err = this.Query().
		Attr("fileId", fileId).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 读取文件的所有片段ID
func (this *FileChunkDAO) FindAllFileChunkIds(fileId int64) ([]int64, error) {
	ones, err := this.Query().
		Attr("fileId", fileId).
		AscPk().
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*FileChunk).Id))
	}
	return result, nil
}

// 删除以前的文件
func (this *FileChunkDAO) DeleteFileChunks(fileId int64) error {
	if fileId <= 0 {
		return errors.New("invalid fileId")
	}
	_, err := this.Query().
		Attr("fileId", fileId).
		Delete()
	return err
}

// 根据ID查找片段
func (this *FileChunkDAO) FindFileChunk(chunkId int64) (*FileChunk, error) {
	one, err := this.Query().
		Pk(chunkId).
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*FileChunk), nil
}
