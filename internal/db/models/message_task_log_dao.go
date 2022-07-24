package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type MessageTaskLogDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedMessageTaskLogDAO.CleanExpiredLogs(nil, 30) // 只保留30天
				if err != nil {
					remotelogs.Error("SharedMessageTaskLogDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

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

// CreateLog 创建日志
func (this *MessageTaskLogDAO) CreateLog(tx *dbs.Tx, taskId int64, isOk bool, errMsg string, response string) error {
	var op = NewMessageTaskLogOperator()
	op.TaskId = taskId
	op.IsOk = isOk
	op.Error = errMsg
	op.Response = response
	op.Day = timeutil.Format("Ymd")
	return this.Save(tx, op)
}

// CountLogs 计算日志数量
func (this *MessageTaskLogDAO) CountLogs(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		Where("taskId IN (SELECT id FROM " + SharedMessageTaskDAO.Table + ")").
		Count()
}

// ListLogs 列出单页日志
func (this *MessageTaskLogDAO) ListLogs(tx *dbs.Tx, offset int64, size int64) (result []*MessageTaskLog, err error) {
	_, err = this.Query(tx).
		Where("taskId IN (SELECT id FROM " + SharedMessageTaskDAO.Table + ")").
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CleanExpiredLogs 清理
func (this *MessageTaskLogDAO) CleanExpiredLogs(tx *dbs.Tx, days int) error {
	if days <= 0 {
		days = 30
	}
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Where("(day IS NULL OR day<:day)").
		Param("day", day).
		Delete()
	return err
}
