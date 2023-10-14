package models

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
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

// CreateMessageTask 创建任务
func (this *MessageTaskDAO) CreateMessageTask(tx *dbs.Tx, recipientId int64, instanceId int64, user string, subject string, body string, isPrimary bool) (int64, error) {
	if !teaconst.IsPlus {
		return 0, nil
	}

	var hash = stringutil.Md5(types.String(recipientId) + "@" + types.String(instanceId) + "@" + user + "@" + subject + "@" + types.String(isPrimary))
	recipientInstanceId, err := SharedMessageRecipientDAO.FindRecipientInstanceId(tx, recipientId)
	if err != nil {
		return 0, err
	}
	if recipientInstanceId > 0 {
		hashLifeSeconds, err := SharedMessageMediaInstanceDAO.FindInstanceHashLifeSeconds(tx, recipientInstanceId)
		if err != nil {
			return 0, err
		}

		if hashLifeSeconds >= 0 { // 意味着此值如果小于0，则不做判断
			lastMessageAt, err := this.Query(tx).
				Attr("hash", hash).
				Result("createdAt").
				DescPk().
				FindInt64Col(0)
			if err != nil {
				return 0, err
			}

			// 对于同一个人N分钟内消息不重复发送
			if hashLifeSeconds <= 0 {
				hashLifeSeconds = 60
			}
			if lastMessageAt > 0 && time.Now().Unix()-lastMessageAt < int64(hashLifeSeconds) {
				return 0, nil
			}
		}
	}

	var op = NewMessageTaskOperator()
	op.RecipientId = recipientId
	op.InstanceId = instanceId
	op.Hash = hash
	op.User = user
	op.Subject = subject
	op.Body = body
	op.IsPrimary = isPrimary
	op.Day = timeutil.Format("Ymd")
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
		Where("(recipientId=0 OR recipientId IN (SELECT id FROM "+SharedMessageRecipientDAO.Table+" WHERE state=1 AND isOn=1 AND (timeFrom IS NULL OR timeTo IS NULL OR :time BETWEEN timeFrom AND timeTo)))").
		Param("time", timeutil.Format("H:i:s")).
		Desc("isPrimary").
		AscPk().
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// CountMessageTasksWithStatus 根据状态计算任务数量
func (this *MessageTaskDAO) CountMessageTasksWithStatus(tx *dbs.Tx, status MessageTaskStatus) (int64, error) {
	return this.Query(tx).
		State(MessageTaskStateEnabled).
		Attr("status", status).
		Count()
}

// ListMessageTasksWithStatus 根据状态列出单页任务
func (this *MessageTaskDAO) ListMessageTasksWithStatus(tx *dbs.Tx, status MessageTaskStatus, offset int64, size int64) (result []*MessageTask, err error) {
	_, err = this.Query(tx).
		State(MessageTaskStateEnabled).
		Attr("status", status).
		Desc("isPrimary").
		AscPk().
		Offset(offset).
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
	var op = NewMessageTaskOperator()
	op.Id = taskId
	op.Status = status
	op.SentAt = time.Now().Unix()
	if len(result) > 0 {
		op.Result = result
	}
	return this.Save(tx, op)
}

// CreateMessageTasks 从集群、节点或者服务中创建任务
func (this *MessageTaskDAO) CreateMessageTasks(tx *dbs.Tx, role nodeconfigs.NodeRole, clusterId int64, nodeId int64, serverId int64, messageType MessageType, subject string, body string) error {
	if !teaconst.IsPlus {
		return nil
	}

	receivers, err := SharedMessageReceiverDAO.FindEnabledBestFitReceivers(tx, role, clusterId, nodeId, serverId, messageType)
	if err != nil {
		return err
	}
	var allRecipientIds = []int64{}
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

	var sentMap = map[int64]bool{} // recipientId => bool 用来检查是否已经发送，防止重复发送给某个接收人
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
