package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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

var SharedFileDAO *FileDAO

func init() {
	dbs.OnReady(func() {
		SharedFileDAO = NewFileDAO()
	})
}

// 启用条目
func (this *FileDAO) EnableFile(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", FileStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *FileDAO) DisableFile(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", FileStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *FileDAO) FindEnabledFile(tx *dbs.Tx, id int64) (*File, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", FileStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*File), err
}

// 创建文件
func (this *FileDAO) CreateFile(tx *dbs.Tx, adminId int64, userId int64, businessType, description string, filename string, size int64, isPublic bool) (int64, error) {
	op := NewFileOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.Type = businessType
	op.Description = description
	op.State = FileStateEnabled
	op.Size = size
	op.Filename = filename
	op.IsPublic = isPublic
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// 将文件置为已完成
func (this *FileDAO) UpdateFileIsFinished(tx *dbs.Tx, fileId int64) error {
	_, err := this.Query(tx).
		Pk(fileId).
		Set("isFinished", true).
		Update()
	return err
}
