package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

type MessageTaskStatus = int

const (
	MessageTaskStateEnabled  = 1 // 已启用
	MessageTaskStateDisabled = 0 // 已禁用

	MessageTaskStatusNone    MessageTaskStatus = 0 // 普通状态
	MessageTaskStatusSending MessageTaskStatus = 1 // 发送中
	MessageTaskStatusSuccess MessageTaskStatus = 2 // 发送成功
	MessageTaskStatusFailed  MessageTaskStatus = 3 // 发送失败
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

// CreateMessageTask 创建任务
func (this *MessageTaskDAO) CreateMessageTask(tx *dbs.Tx, recipientId int64, instanceId int64, user string, subject string, body string, isPrimary bool) (int64, error) {
	op := NewMessageTaskOperator()
	op.RecipientId = recipientId
	op.InstanceId = instanceId
	op.User = user
	op.Subject = subject
	op.Body = body
	op.IsPrimary = isPrimary
	op.Status = MessageTaskStatusNone
	op.State = MessageTaskStateEnabled
	return this.SaveInt64(tx, op)
}

// FindSendingMessageTasks 查找需要发送的任务
func (this *MessageTaskDAO) FindSendingMessageTasks(tx *dbs.Tx, size int64) (result []*MessageTask, err error) {
	if size <= 0 {
		return nil, nil
	}
	_, err = this.Query(tx).
		State(MessageTaskStateEnabled).
		Attr("status", MessageTaskStatusNone).
		Desc("isPrimary").
		AscPk().
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// UpdateMessageTaskStatus 设置发送的状态
func (this *MessageTaskDAO) UpdateMessageTaskStatus(tx *dbs.Tx, taskId int64, status MessageTaskStatus, result []byte) error {
	if taskId <= 0 {
		return errors.New("invalid taskId")
	}
	op := NewMessageTaskOperator()
	op.Id = taskId
	op.Status = status
	op.SentAt = time.Now().Unix()
	if len(result) > 0 {
		op.Result = result
	}
	return this.Save(tx, op)
}

// CreateMessageTasks 从集群、节点或者服务中创建任务
func (this *MessageTaskDAO) CreateMessageTasks(tx *dbs.Tx, target MessageTaskTarget, messageType MessageType, subject string, body string) error {
	receivers, err := SharedMessageReceiverDAO.FindAllEnabledReceivers(tx, target, messageType)
	if err != nil {
		return err
	}
	allRecipientIds := []int64{}
	for _, receiver := range receivers {
		if receiver.RecipientId > 0 {
			allRecipientIds = append(allRecipientIds, int64(receiver.RecipientId))
		} else if receiver.RecipientGroupId > 0 {
			recipientIds, err := SharedMessageRecipientDAO.FindAllEnabledAndOnRecipientIdsWithGroup(tx, int64(receiver.RecipientGroupId))
			if err != nil {
				return err
			}
			allRecipientIds = append(allRecipientIds, recipientIds...)
		}
	}

	sentMap := map[int64]bool{} // recipientId => bool 用来检查是否已经发送，防止重复发送给某个接收人
	for _, recipientId := range allRecipientIds {
		_, ok := sentMap[recipientId]
		if ok {
			continue
		}
		sentMap[recipientId] = true
		_, err := this.CreateMessageTask(tx, recipientId, 0, "", subject, body, false)
		if err != nil {
			return err
		}
	}

	return nil
}
