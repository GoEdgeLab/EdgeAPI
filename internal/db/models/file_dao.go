package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"mime/multipart"
	"os"
)

const (
	FileStateEnabled  = 1 // 已启用
	FileStateDisabled = 0 // 已禁用
)

type FileDAO dbs.DAO

func NewFileDAO() *FileDAO {
	return dbs.NewDAO(&FileDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeFiles",
			Model:  new(File),
			PkName: "id",
		},
	}).(*FileDAO)
}

var SharedFileDAO = NewFileDAO()

// 启用条目
func (this *FileDAO) EnableFile(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", FileStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *FileDAO) DisableFile(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", FileStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *FileDAO) FindEnabledFile(id int64) (*File, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", FileStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*File), err
}

// 创建文件
func (this *FileDAO) CreateFileFromReader(businessType, description string, filename string, body *multipart.FileHeader, order int) (int, error) {
	file, err := body.Open()
	if err != nil {
		return 0, err
	}

	op := NewFileOperator()
	op.Type = businessType
	op.Description = description
	op.State = FileStateEnabled
	op.Size = body.Size
	op.Order = order
	op.Filename = filename
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}

	fileId := types.Int(op.Id)

	// 保存chunk
	buf := make([]byte, 512*1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			err1 := SharedFileChunkDAO.CreateFileChunk(fileId, buf[:n])
			if err1 != nil {
				return 0, err1
			}
		}
		if err != nil {
			break
		}
	}

	return fileId, nil
}

// 创建一个空文件
func (this *FileDAO) UploadLocalFile(businessType string, localFile string, filename string) (fileId int, err error) {
	reader, err := os.Open(localFile)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = reader.Close()
	}()
	stat, err := reader.Stat()
	if err != nil {
		return 0, err
	}

	op := NewFileOperator()
	op.Type = businessType
	op.Filename = filename
	op.Size = stat.Size()
	op.State = FileStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return
	}
	fileId = types.Int(op.Id)

	buf := make([]byte, 512*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			err1 := SharedFileChunkDAO.CreateFileChunk(fileId, buf[:n])
			if err1 != nil {
				return 0, err1
			}
		}
		if err != nil {
			break
		}
	}

	return fileId, nil
}
