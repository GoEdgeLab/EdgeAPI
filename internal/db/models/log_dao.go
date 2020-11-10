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

var SharedLogDAO *LogDAO

func init() {
	dbs.OnReady(func() {
		SharedLogDAO = NewLogDAO()
	})
}

// 创建管理员日志
func (this *LogDAO) CreateLog(adminType string, adminId int64, level string, description string, action string, ip string) error {
	op := NewLogOperator()
	op.Type = adminType
	op.AdminId, op.Level, op.Description, op.Action, op.Ip = adminId, level, description, action, ip
	op.Type = LogTypeAdmin
	_, err := this.Save(op)
	return err
}

// 计算所有日志数量
func (this *LogDAO) CountAllLogs() (int64, error) {
	return this.Query().
		Count()
}

// 列出单页日志
func (this *LogDAO) ListLogs(offset int64, size int64) (result []*Log, err error) {
	_, err = this.Query().
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}
