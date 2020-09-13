package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	OriginServerStateEnabled  = 1 // 已启用
	OriginServerStateDisabled = 0 // 已禁用
)

type OriginServerDAO dbs.DAO

func NewOriginServerDAO() *OriginServerDAO {
	return dbs.NewDAO(&OriginServerDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeOriginServers",
			Model:  new(OriginServer),
			PkName: "id",
		},
	}).(*OriginServerDAO)
}

var SharedOriginServerDAO = NewOriginServerDAO()

// 启用条目
func (this *OriginServerDAO) EnableOriginServer(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", OriginServerStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *OriginServerDAO) DisableOriginServer(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", OriginServerStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *OriginServerDAO) FindEnabledOriginServer(id int64) (*OriginServer, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", OriginServerStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*OriginServer), err
}

// 根据主键查找名称
func (this *OriginServerDAO) FindOriginServerName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建源站
func (this *OriginServerDAO) CreateOriginServer(name string, addrJSON string, description string) (originId int64, err error) {
	op := NewOriginServerOperator()
	op.IsOn = true
	op.Name = name
	op.Addr = addrJSON
	op.Description = description
	op.State = OriginServerStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return
	}
	return types.Int64(op.Id), nil
}

// 修改源站
func (this *OriginServerDAO) UpdateOriginServer(originId int64, name string, addrJSON string, description string) error {
	if originId <= 0 {
		return errors.New("invalid originId")
	}
	op := NewOriginServerOperator()
	op.Id = originId
	op.Name = name
	op.Addr = addrJSON
	op.Description = description
	op.Version = dbs.SQL("version+1")
	_, err := this.Save(op)
	return err
}
