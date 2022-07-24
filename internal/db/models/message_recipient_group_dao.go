package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	MessageRecipientGroupStateEnabled  = 1 // 已启用
	MessageRecipientGroupStateDisabled = 0 // 已禁用
)

type MessageRecipientGroupDAO dbs.DAO

func NewMessageRecipientGroupDAO() *MessageRecipientGroupDAO {
	return dbs.NewDAO(&MessageRecipientGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMessageRecipientGroups",
			Model:  new(MessageRecipientGroup),
			PkName: "id",
		},
	}).(*MessageRecipientGroupDAO)
}

var SharedMessageRecipientGroupDAO *MessageRecipientGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedMessageRecipientGroupDAO = NewMessageRecipientGroupDAO()
	})
}

// 启用条目
func (this *MessageRecipientGroupDAO) EnableMessageRecipientGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageRecipientGroupStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *MessageRecipientGroupDAO) DisableMessageRecipientGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageRecipientGroupStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *MessageRecipientGroupDAO) FindEnabledMessageRecipientGroup(tx *dbs.Tx, id int64) (*MessageRecipientGroup, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", MessageRecipientGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MessageRecipientGroup), err
}

// 根据主键查找名称
func (this *MessageRecipientGroupDAO) FindMessageRecipientGroupName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建分组
func (this *MessageRecipientGroupDAO) CreateGroup(tx *dbs.Tx, name string) (int64, error) {
	var op = NewMessageRecipientGroupOperator()
	op.Name = name
	op.IsOn = true
	op.State = MessageRecipientStateEnabled
	return this.SaveInt64(tx, op)
}

// 修改分组
func (this *MessageRecipientGroupDAO) UpdateGroup(tx *dbs.Tx, groupId int64, name string, isOn bool) error {
	if groupId <= 0 {
		return errors.New("invalid groupId")
	}
	var op = NewMessageRecipientGroupOperator()
	op.Id = groupId
	op.Name = name
	op.IsOn = isOn
	return this.Save(tx, op)
}

// 查找所有分组
func (this *MessageRecipientGroupDAO) FindAllEnabledGroups(tx *dbs.Tx) (result []*MessageRecipientGroup, err error) {
	_, err = this.Query(tx).
		State(MessageRecipientStateEnabled).
		Slice(&result).
		Desc("order").
		AscPk().
		FindAll()
	return
}
