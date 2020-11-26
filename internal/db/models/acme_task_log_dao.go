package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type ACMETaskLogDAO dbs.DAO

func NewACMETaskLogDAO() *ACMETaskLogDAO {
	return dbs.NewDAO(&ACMETaskLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeACMETaskLogs",
			Model:  new(ACMETaskLog),
			PkName: "id",
		},
	}).(*ACMETaskLogDAO)
}

var SharedACMETaskLogDAO *ACMETaskLogDAO

func init() {
	dbs.OnReady(func() {
		SharedACMETaskLogDAO = NewACMETaskLogDAO()
	})
}

// 生成日志
func (this *ACMETaskLogDAO) CreateACMELog(taskId int64, isOk bool, errMsg string) error {
	op := NewACMETaskLogOperator()
	op.TaskId = taskId
	op.Error = errMsg
	op.IsOk = isOk
	_, err := this.Save(op)
	return err
}
