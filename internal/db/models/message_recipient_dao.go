package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	MessageRecipientStateEnabled  = 1 // 已启用
	MessageRecipientStateDisabled = 0 // 已禁用
)

type MessageRecipientDAO dbs.DAO

func NewMessageRecipientDAO() *MessageRecipientDAO {
	return dbs.NewDAO(&MessageRecipientDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMessageRecipients",
			Model:  new(MessageRecipient),
			PkName: "id",
		},
	}).(*MessageRecipientDAO)
}

var SharedMessageRecipientDAO *MessageRecipientDAO

func init() {
	dbs.OnReady(func() {
		SharedMessageRecipientDAO = NewMessageRecipientDAO()
	})
}

// 启用条目
func (this *MessageRecipientDAO) EnableMessageRecipient(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageRecipientStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *MessageRecipientDAO) DisableMessageRecipient(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageRecipientStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *MessageRecipientDAO) FindEnabledMessageRecipient(tx *dbs.Tx, id int64) (*MessageRecipient, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", MessageRecipientStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MessageRecipient), err
}

// 创建接收人
func (this *MessageRecipientDAO) CreateRecipient(tx *dbs.Tx, adminId int64, instanceId int64, user string, groupIds []int64, description string) (int64, error) {
	op := NewMessageRecipientOperator()
	op.AdminId = adminId
	op.InstanceId = instanceId
	op.User = user
	op.Description = description

	// 分组
	if len(groupIds) == 0 {
		groupIds = []int64{}
	}
	groupIdsJSON, err := json.Marshal(groupIds)
	if err != nil {
		return 0, err
	}
	op.GroupIds = groupIdsJSON

	op.IsOn = true
	op.State = MessageRecipientStateEnabled
	return this.SaveInt64(tx, op)
}

// 修改接收人
func (this *MessageRecipientDAO) UpdateRecipient(tx *dbs.Tx, recipientId int64, adminId int64, instanceId int64, user string, groupIds []int64, description string, isOn bool) error {
	if recipientId <= 0 {
		return errors.New("invalid recipientId")
	}

	op := NewMessageRecipientOperator()
	op.Id = recipientId
	op.AdminId = adminId
	op.InstanceId = instanceId
	op.User = user

	// 分组
	if len(groupIds) == 0 {
		groupIds = []int64{}
	}
	groupIdsJSON, err := json.Marshal(groupIds)
	if err != nil {
		return err
	}
	op.GroupIds = groupIdsJSON

	op.Description = description
	op.IsOn = isOn
	return this.Save(tx, op)
}

// 计算接收人数量
func (this *MessageRecipientDAO) CountAllEnabledRecipients(tx *dbs.Tx, adminId int64, groupId int64, mediaType string, keyword string) (int64, error) {
	query := this.Query(tx)
	if adminId > 0 {
		query.Attr("adminId", adminId)
	}
	if groupId > 0 {
		query.Attr("groupId", groupId)
	}
	if len(mediaType) > 0 {
		query.Where("instanceId IN (SELECT id FROM "+SharedMessageMediaInstanceDAO.Table+" WHERE state=1 AND mediaType=:mediaType)").
			Param("mediaType", mediaType)
	}
	if len(keyword) > 0 {
		query.Where("(`user` LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	return query.
		State(MessageRecipientStateEnabled).
		Where("adminId IN (SELECT id FROM " + SharedAdminDAO.Table + " WHERE state=1)").
		Where("instanceId IN (SELECT id FROM " + SharedMessageMediaInstanceDAO.Table + " WHERE state=1)").
		Count()
}

// 列出单页接收人
func (this *MessageRecipientDAO) ListAllEnabledRecipients(tx *dbs.Tx, adminId int64, groupId int64, mediaType string, keyword string, offset int64, size int64) (result []*MessageRecipient, err error) {
	query := this.Query(tx)
	if adminId > 0 {
		query.Attr("adminId", adminId)
	}
	if groupId > 0 {
		query.Attr("groupId", groupId)
	}
	if len(mediaType) > 0 {
		query.Where("instanceId IN (SELECT id FROM "+SharedMessageMediaInstanceDAO.Table+" WHERE state=1 AND mediaType=:mediaType)").
			Param("mediaType", mediaType)
	}
	if len(keyword) > 0 {
		query.Where("(`user` LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	_, err = query.
		State(MessageRecipientStateEnabled).
		Where("adminId IN (SELECT id FROM " + SharedAdminDAO.Table + " WHERE state=1)").
		Where("instanceId IN (SELECT id FROM " + SharedMessageMediaInstanceDAO.Table + " WHERE state=1)").
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}
