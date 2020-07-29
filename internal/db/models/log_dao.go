package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type LogDAO dbs.DAO

func NewLogDAO() *LogDAO {
	return dbs.NewDAO(&LogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeLogs",
			Model:  new(Log),
			PkName: "id",
		},
	}).(*LogDAO)
}

var SharedLogDAO = NewLogDAO()

// 创建管理员日志
func (this *LogDAO) CreateAdminLog(adminId int64, level string, description string, action string, ip string) error {
	op := NewLogOperator()
	op.AdminId, op.Level, op.Description, op.Action, op.Ip = adminId, level, description, action, ip
	op.Type = LogTypeAdmin
	_, err := this.Save(op)
	return err
}
