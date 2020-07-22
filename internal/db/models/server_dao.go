package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ServerStateEnabled  = 1 // 已启用
	ServerStateDisabled = 0 // 已禁用
)

type ServerDAO dbs.DAO

func NewServerDAO() *ServerDAO {
	return dbs.NewDAO(&ServerDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServers",
			Model:  new(Server),
			PkName: "id",
		},
	}).(*ServerDAO)
}

var SharedServerDAO = NewServerDAO()

// 启用条目
func (this *ServerDAO) EnableServer(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ServerStateEnabled).
		Update()
}

// 禁用条目
func (this *ServerDAO) DisableServer(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ServerStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *ServerDAO) FindEnabledServer(id uint32) (*Server, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ServerStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Server), err
}
