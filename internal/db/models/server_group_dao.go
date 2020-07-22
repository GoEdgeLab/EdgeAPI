package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ServerGroupStateEnabled  = 1 // 已启用
	ServerGroupStateDisabled = 0 // 已禁用
)

type ServerGroupDAO dbs.DAO

func NewServerGroupDAO() *ServerGroupDAO {
	return dbs.NewDAO(&ServerGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerGroups",
			Model:  new(ServerGroup),
			PkName: "id",
		},
	}).(*ServerGroupDAO)
}

var SharedServerGroupDAO = NewServerGroupDAO()

// 启用条目
func (this *ServerGroupDAO) EnableServerGroup(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ServerGroupStateEnabled).
		Update()
}

// 禁用条目
func (this *ServerGroupDAO) DisableServerGroup(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ServerGroupStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *ServerGroupDAO) FindEnabledServerGroup(id uint32) (*ServerGroup, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ServerGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ServerGroup), err
}

// 根据主键查找名称
func (this *ServerGroupDAO) FindServerGroupName(id uint32) (string, error) {
	name, err := this.Query().
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}
