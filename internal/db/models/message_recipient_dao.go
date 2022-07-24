package models

import (
	"encoding/json"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"regexp"
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

// EnableMessageRecipient 启用条目
func (this *MessageRecipientDAO) EnableMessageRecipient(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageRecipientStateEnabled).
		Update()
	return err
}

// DisableMessageRecipient 禁用条目
func (this *MessageRecipientDAO) DisableMessageRecipient(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageRecipientStateDisabled).
		Update()
	return err
}

// FindEnabledMessageRecipient 查找启用中的条目
func (this *MessageRecipientDAO) FindEnabledMessageRecipient(tx *dbs.Tx, recipientId int64, cacheMap *utils.CacheMap,
) (*MessageRecipient, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":record:" + types.String(recipientId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*MessageRecipient), nil
	}

	result, err := this.Query(tx).
		Pk(recipientId).
		Attr("state", MessageRecipientStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, result)
	}

	return result.(*MessageRecipient), err
}

// CreateRecipient 创建接收人
func (this *MessageRecipientDAO) CreateRecipient(tx *dbs.Tx, adminId int64, instanceId int64, user string, groupIds []int64, description string, timeFrom string, timeTo string) (int64, error) {
	var op = NewMessageRecipientOperator()
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

	// 判断格式
	var timeReg = regexp.MustCompile(`^\d+:\d+:\d+$`)
	if timeReg.MatchString(timeFrom) {
		op.TimeFrom = timeFrom
	}
	if timeReg.MatchString(timeTo) {
		op.TimeTo = timeTo
	}

	op.IsOn = true
	op.State = MessageRecipientStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateRecipient 修改接收人
func (this *MessageRecipientDAO) UpdateRecipient(tx *dbs.Tx, recipientId int64, adminId int64, instanceId int64, user string, groupIds []int64, description string, timeFrom string, timeTo string, isOn bool) error {
	if recipientId <= 0 {
		return errors.New("invalid recipientId")
	}

	var op = NewMessageRecipientOperator()
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

	// 判断格式
	var timeReg = regexp.MustCompile(`^\d+:\d+:\d+$`)
	if timeReg.MatchString(timeFrom) {
		op.TimeFrom = timeFrom
	} else {
		op.TimeFrom = dbs.SQL("NULL")
	}
	if timeReg.MatchString(timeTo) {
		op.TimeTo = timeTo
	} else {
		op.TimeTo = dbs.SQL("NULL")
	}

	op.IsOn = isOn
	return this.Save(tx, op)
}

// CountAllEnabledRecipients 计算接收人数量
func (this *MessageRecipientDAO) CountAllEnabledRecipients(tx *dbs.Tx, adminId int64, groupId int64, mediaType string, keyword string) (int64, error) {
	query := this.Query(tx)
	if adminId > 0 {
		query.Attr("adminId", adminId)
	}
	if groupId > 0 {
		query.Where("JSON_CONTAINS(groupIds, :groupId)").
			Param("groupId", numberutils.FormatInt64(groupId))
	}
	if len(mediaType) > 0 {
		query.Where("instanceId IN (SELECT id FROM "+SharedMessageMediaInstanceDAO.Table+" WHERE state=1 AND mediaType=:mediaType)").
			Param("mediaType", mediaType)
	}
	if len(keyword) > 0 {
		query.Where("(`user` LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	return query.
		State(MessageRecipientStateEnabled).
		Where("adminId IN (SELECT id FROM " + SharedAdminDAO.Table + " WHERE state=1)").
		Where("instanceId IN (SELECT id FROM " + SharedMessageMediaInstanceDAO.Table + " WHERE state=1)").
		Count()
}

// ListAllEnabledRecipients 列出单页接收人
func (this *MessageRecipientDAO) ListAllEnabledRecipients(tx *dbs.Tx, adminId int64, groupId int64, mediaType string, keyword string, offset int64, size int64) (result []*MessageRecipient, err error) {
	query := this.Query(tx)
	if adminId > 0 {
		query.Attr("adminId", adminId)
	}
	if groupId > 0 {
		query.Where("JSON_CONTAINS(groupIds, :groupId)").
			Param("groupId", numberutils.FormatInt64(groupId))
	}
	if len(mediaType) > 0 {
		query.Where("instanceId IN (SELECT id FROM "+SharedMessageMediaInstanceDAO.Table+" WHERE state=1 AND mediaType=:mediaType)").
			Param("mediaType", mediaType)
	}
	if len(keyword) > 0 {
		query.Where("(`user` LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
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

// FindAllEnabledAndOnRecipientIdsWithGroup 查找某个分组下的所有可用接收人ID
func (this *MessageRecipientDAO) FindAllEnabledAndOnRecipientIdsWithGroup(tx *dbs.Tx, groupId int64) ([]int64, error) {
	ones, err := this.Query(tx).
		Where("JSON_CONTAINS(groupIds, :groupId)").
		Param("groupId", numberutils.FormatInt64(groupId)).
		State(MessageRecipientStateEnabled).
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*MessageRecipient).Id))
	}
	return result, nil
}

// FindRecipientInstanceId 查找接收人的媒介
func (this *MessageRecipientDAO) FindRecipientInstanceId(tx *dbs.Tx, recipientId int64) (int64, error) {
	return this.Query(tx).
		Pk(recipientId).
		Result("instanceId").
		FindInt64Col(0)
}
