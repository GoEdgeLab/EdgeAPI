package acme

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
func (this *ACMETaskLogDAO) CreateACMETaskLog(tx *dbs.Tx, taskId int64, isOk bool, errMsg string) error {
	var op = NewACMETaskLogOperator()
	op.TaskId = taskId
	op.Error = errMsg
	op.IsOk = isOk
	err := this.Save(tx, op)
	return err
}

// 取得任务的最后一条执行日志
func (this *ACMETaskLogDAO) FindLatestACMETasKLog(tx *dbs.Tx, taskId int64) (*ACMETaskLog, error) {
	one, err := this.Query(tx).
		Attr("taskId", taskId).
		DescPk().
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*ACMETaskLog), nil
}
