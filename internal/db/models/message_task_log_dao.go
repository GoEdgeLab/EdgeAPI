package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type MessageTaskLogDAO dbs.DAO

func NewMessageTaskLogDAO() *MessageTaskLogDAO {
	return dbs.NewDAO(&MessageTaskLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMessageTaskLogs",
			Model:  new(MessageTaskLog),
			PkName: "id",
		},
	}).(*MessageTaskLogDAO)
}

var SharedMessageTaskLogDAO *MessageTaskLogDAO

func init() {
	dbs.OnReady(func() {
		SharedMessageTaskLogDAO = NewMessageTaskLogDAO()
	})
}

// 创建日志
func (this *MessageTaskLogDAO) CreateLog(tx *dbs.Tx, taskId int64, isOk bool, errMsg string, response string) error {
	op := NewMessageTaskLogOperator()
	op.TaskId = taskId
	op.IsOk = isOk
	op.Error = errMsg
	op.Response = response
	return this.Save(tx, op)
}

// 计算日志数量
func (this *MessageTaskLogDAO) CountLogs(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		Count()
}

// 列出单页日志
func (this *MessageTaskLogDAO) ListLogs(tx *dbs.Tx, offset int64, size int64) (result []*MessageTaskLog, err error) {
	_, err = this.Query(tx).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
