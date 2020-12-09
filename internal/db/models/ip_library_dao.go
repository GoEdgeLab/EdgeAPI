package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	IPLibraryStateEnabled  = 1 // 已启用
	IPLibraryStateDisabled = 0 // 已禁用
)

type IPLibraryDAO dbs.DAO

func NewIPLibraryDAO() *IPLibraryDAO {
	return dbs.NewDAO(&IPLibraryDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeIPLibraries",
			Model:  new(IPLibrary),
			PkName: "id",
		},
	}).(*IPLibraryDAO)
}

var SharedIPLibraryDAO *IPLibraryDAO

func init() {
	dbs.OnReady(func() {
		SharedIPLibraryDAO = NewIPLibraryDAO()
	})
}

// 启用条目
func (this *IPLibraryDAO) EnableIPLibrary(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", IPLibraryStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *IPLibraryDAO) DisableIPLibrary(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", IPLibraryStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *IPLibraryDAO) FindEnabledIPLibrary(id int64) (*IPLibrary, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", IPLibraryStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*IPLibrary), err
}

// 查找某个类型的IP库列表
func (this *IPLibraryDAO) FindAllEnabledIPLibrariesWithType(libraryType string) (result []*IPLibrary, err error) {
	_, err = this.Query().
		State(IPLibraryStateEnabled).
		Attr("type", libraryType).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 查找某个类型的最新的IP库
func (this *IPLibraryDAO) FindLatestIPLibraryWithType(libraryType string) (*IPLibrary, error) {
	one, err := this.Query().
		State(IPLibraryStateEnabled).
		Attr("type", libraryType).
		DescPk().
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*IPLibrary), nil
}

// 创建新的IP库
func (this *IPLibraryDAO) CreateIPLibrary(libraryType string, fileId int64) (int64, error) {
	op := NewIPLibraryOperator()
	op.Type = libraryType
	op.FileId = fileId
	op.State = IPLibraryStateEnabled
	err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}
