package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ipconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	IPListStateEnabled  = 1 // 已启用
	IPListStateDisabled = 0 // 已禁用
)

type IPListDAO dbs.DAO

func NewIPListDAO() *IPListDAO {
	return dbs.NewDAO(&IPListDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeIPLists",
			Model:  new(IPList),
			PkName: "id",
		},
	}).(*IPListDAO)
}

var SharedIPListDAO *IPListDAO

func init() {
	dbs.OnReady(func() {
		SharedIPListDAO = NewIPListDAO()
	})
}

// 启用条目
func (this *IPListDAO) EnableIPList(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", IPListStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *IPListDAO) DisableIPList(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", IPListStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *IPListDAO) FindEnabledIPList(id int64) (*IPList, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", IPListStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*IPList), err
}

// 根据主键查找名称
func (this *IPListDAO) FindIPListName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建名单
func (this *IPListDAO) CreateIPList(listType ipconfigs.IPListType, name string, code string, timeoutJSON []byte) (int64, error) {
	op := NewIPListOperator()
	op.IsOn = true
	op.State = IPListStateEnabled
	op.Type = listType
	op.Name = name
	op.Code = code
	if len(timeoutJSON) > 0 {
		op.Timeout = timeoutJSON
	}
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改名单
func (this *IPListDAO) UpdateIPList(listId int64, name string, code string, timeoutJSON []byte) error {
	if listId <= 0 {
		return errors.New("invalid listId")
	}
	op := NewIPListOperator()
	op.Id = listId
	op.Name = name
	op.Code = code
	if len(timeoutJSON) > 0 {
		op.Timeout = timeoutJSON
	} else {
		op.Timeout = "null"
	}
	_, err := this.Save(op)
	return err
}

// 增加版本
func (this *IPListDAO) IncreaseVersion(listId int64) (int64, error) {
	if listId <= 0 {
		return 0, errors.New("invalid listId")
	}
	op := NewIPListOperator()
	op.Id = listId
	op.Version = dbs.SQL("version+1")
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}

	return this.Query().
		Pk(listId).
		Result("version").
		FindInt64Col(0)
}
