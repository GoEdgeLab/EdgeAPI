package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	ServerGroupStateEnabled  = 1 // 已启用
	ServerGroupStateDisabled = 0 // 已禁用
)

type ServerGroupDAO dbs.DAO

func NewServerGroupDAO() *ServerGroupDAO {
	return dbs.NewDAO(&ServerGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerGroups",
			Model:  new(ServerGroup),
			PkName: "id",
		},
	}).(*ServerGroupDAO)
}

var SharedServerGroupDAO *ServerGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedServerGroupDAO = NewServerGroupDAO()
	})
}

// 启用条目
func (this *ServerGroupDAO) EnableServerGroup(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ServerGroupStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *ServerGroupDAO) DisableServerGroup(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ServerGroupStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *ServerGroupDAO) FindEnabledServerGroup(id int64) (*ServerGroup, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ServerGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ServerGroup), err
}

// 根据主键查找名称
func (this *ServerGroupDAO) FindServerGroupName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建分组
func (this *ServerGroupDAO) CreateGroup(name string) (groupId int64, err error) {
	op := NewServerGroupOperator()
	op.State = ServerGroupStateEnabled
	op.Name = name
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改分组
func (this *ServerGroupDAO) UpdateGroup(groupId int64, name string) error {
	if groupId <= 0 {
		return errors.New("invalid groupId")
	}
	op := NewServerGroupOperator()
	op.Id = groupId
	op.Name = name
	_, err := this.Save(op)
	return err
}

// 查找所有分组
func (this *ServerGroupDAO) FindAllEnabledGroups() (result []*ServerGroup, err error) {
	_, err = this.Query().
		State(ServerGroupStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 修改分组排序
func (this *ServerGroupDAO) UpdateGroupOrders(groupIds []int64) error {
	for index, groupId := range groupIds {
		_, err := this.Query().
			Pk(groupId).
			Set("order", len(groupIds)-index).
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}
