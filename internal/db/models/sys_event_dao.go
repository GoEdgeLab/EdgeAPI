package models

import (
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type SysEventDAO dbs.DAO

func NewSysEventDAO() *SysEventDAO {
	return dbs.NewDAO(&SysEventDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeSysEvents",
			Model:  new(SysEvent),
			PkName: "id",
		},
	}).(*SysEventDAO)
}

var SharedSysEventDAO *SysEventDAO

func init() {
	dbs.OnReady(func() {
		SharedSysEventDAO = NewSysEventDAO()
	})
}

// 创建事件
func (this *SysEventDAO) CreateEvent(tx *dbs.Tx, event EventInterface) error {
	if event == nil {
		return errors.New("event should not be nil")
	}

	var op = NewSysEventOperator()
	op.Type = event.Type()

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}
	op.Params = eventJSON

	err = this.Save(tx, op)
	return err
}

// 查找事件
func (this *SysEventDAO) FindEvents(tx *dbs.Tx, size int64) (result []*SysEvent, err error) {
	_, err = this.Query(tx).
		AscPk().
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// 删除事件
func (this *SysEventDAO) DeleteEvent(tx *dbs.Tx, eventId int64) error {
	_, err := this.Query(tx).
		Pk(eventId).
		Delete()
	return err
}
