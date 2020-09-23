package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	DBNodeStateEnabled  = 1 // 已启用
	DBNodeStateDisabled = 0 // 已禁用
)

type DBNodeDAO dbs.DAO

func NewDBNodeDAO() *DBNodeDAO {
	return dbs.NewDAO(&DBNodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeDBNodes",
			Model:  new(DBNode),
			PkName: "id",
		},
	}).(*DBNodeDAO)
}

var SharedDBNodeDAO = NewDBNodeDAO()

// 启用条目
func (this *DBNodeDAO) EnableDBNode(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", DBNodeStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *DBNodeDAO) DisableDBNode(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", DBNodeStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *DBNodeDAO) FindEnabledDBNode(id uint32) (*DBNode, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", DBNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*DBNode), err
}

// 根据主键查找名称
func (this *DBNodeDAO) FindDBNodeName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
