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

const (
	MessageTaskStateEnabled  = 1 // 已启用
	MessageTaskStateDisabled = 0 // 已禁用
)

type MessageTaskDAO dbs.DAO

func NewMessageTaskDAO() *MessageTaskDAO {
	return dbs.NewDAO(&MessageTaskDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMessageTasks",
			Model:  new(MessageTask),
			PkName: "id",
		},
	}).(*MessageTaskDAO)
}

var SharedMessageTaskDAO *MessageTaskDAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedMessageTaskDAO.CleanExpiredMessageTasks(nil, 30) // 只保留30天
				if err != nil {
					remotelogs.Error("SharedMessageTaskDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

func init() {
	dbs.OnReady(func() {
		SharedMessageTaskDAO = NewMessageTaskDAO()
	})
}

// EnableMessageTask 启用条目
func (this *MessageTaskDAO) EnableMessageTask(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageTaskStateEnabled).
		Update()
	return err
}

// DisableMessageTask 禁用条目
func (this *MessageTaskDAO) DisableMessageTask(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageTaskStateDisabled).
		Update()
	return err
}

// FindEnabledMessageTask 查找启用中的条目
func (this *MessageTaskDAO) FindEnabledMessageTask(tx *dbs.Tx, id int64) (*MessageTask, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", MessageTaskStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MessageTask), err
}

// CleanExpiredMessageTasks 清理
func (this *MessageTaskDAO) CleanExpiredMessageTasks(tx *dbs.Tx, days int) error {
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
