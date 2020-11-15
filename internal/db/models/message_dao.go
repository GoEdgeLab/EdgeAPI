package models

import (
	"crypto/md5"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

const (
	MessageStateEnabled  = 1 // 已启用
	MessageStateDisabled = 0 // 已禁用

	MessageLevelInfo    = "info"
	MessageLevelWarning = "warning"
	MessageLevelError   = "error"
)

type MessageType = string

const (
	MessageTypeHealthCheckFailed    MessageType = "HealthCheckFailed"
	MessageTypeNodeInactive         MessageType = "NodeInactive"
	MessageTypeClusterDNSSyncFailed MessageType = "ClusterDNSSyncFailed"
)

type MessageDAO dbs.DAO

func NewMessageDAO() *MessageDAO {
	return dbs.NewDAO(&MessageDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMessages",
			Model:  new(Message),
			PkName: "id",
		},
	}).(*MessageDAO)
}

var SharedMessageDAO *MessageDAO

func init() {
	dbs.OnReady(func() {
		SharedMessageDAO = NewMessageDAO()
	})
}

// 启用条目
func (this *MessageDAO) EnableMessage(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", MessageStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *MessageDAO) DisableMessage(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", MessageStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *MessageDAO) FindEnabledMessage(id int64) (*Message, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", MessageStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Message), err
}

// 创建集群消息
func (this *MessageDAO) CreateClusterMessage(clusterId int64, messageType MessageType, level string, body string, paramsJSON []byte) error {
	_, err := this.createMessage(clusterId, 0, messageType, level, body, paramsJSON)
	return err
}

// 创建节点消息
func (this *MessageDAO) CreateNodeMessage(clusterId int64, nodeId int64, messageType MessageType, level string, body string, paramsJSON []byte) error {
	_, err := this.createMessage(clusterId, nodeId, messageType, level, body, paramsJSON)
	return err
}

// 删除某天之前的消息
func (this *MessageDAO) DeleteMessagesBeforeDay(dayTime time.Time) error {
	day := timeutil.Format("Ymd", dayTime)
	_, err := this.Query().
		Where("day<:day").
		Param("day", day).
		Delete()
	return err
}

// 计算未读消息数量
func (this *MessageDAO) CountUnreadMessages() (int64, error) {
	return this.Query().
		Attr("isRead", false).
		Count()
}

// 列出单页未读消息
func (this *MessageDAO) ListUnreadMessages(offset int64, size int64) (result []*Message, err error) {
	_, err = this.Query().
		Attr("isRead", false).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 设置消息已读状态
func (this *MessageDAO) UpdateMessageRead(messageId int64, b bool) error {
	if messageId <= 0 {
		return errors.New("invalid messageId")
	}
	op := NewMessageOperator()
	op.Id = messageId
	op.IsRead = b
	_, err := this.Save(op)
	return err
}

// 设置一组消息为已读状态
func (this *MessageDAO) UpdateMessagesRead(messageIds []int64, b bool) error {
	// 这里我们一个一个更改，因为In语句不容易Prepare，且效率不高
	for _, messageId := range messageIds {
		err := this.UpdateMessageRead(messageId, b)
		if err != nil {
			return err
		}
	}
	return nil
}

// 设置所有消息为已读
func (this *MessageDAO) UpdateAllMessagesRead() error {
	_, err := this.Query().
		Attr("isRead", false).
		Set("isRead", true).
		Update()
	return err
}

// 创建消息
func (this *MessageDAO) createMessage(clusterId int64, nodeId int64, messageType MessageType, level string, body string, paramsJSON []byte) (int64, error) {
	h := md5.New()
	h.Write([]byte(body))
	h.Write(paramsJSON)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// TODO 检查同样的消息最近是否发送过

	// 创建新消息
	op := NewMessageOperator()
	op.AdminId = 0 // TODO
	op.UserId = 0  // TODO
	op.ClusterId = clusterId
	op.NodeId = nodeId
	op.Type = messageType
	op.Level = level
	op.Body = body
	if len(paramsJSON) > 0 {
		op.Params = paramsJSON
	}
	op.IsRead = false
	op.State = MessageStateEnabled
	op.CreatedAt = time.Now().Unix()
	op.Day = timeutil.Format("Ymd")
	op.Hash = hash

	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}
