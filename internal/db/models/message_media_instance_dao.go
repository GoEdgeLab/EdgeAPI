package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

const (
	MessageMediaInstanceStateEnabled  = 1 // 已启用
	MessageMediaInstanceStateDisabled = 0 // 已禁用
)

type MessageMediaInstanceDAO dbs.DAO

func NewMessageMediaInstanceDAO() *MessageMediaInstanceDAO {
	return dbs.NewDAO(&MessageMediaInstanceDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMessageMediaInstances",
			Model:  new(MessageMediaInstance),
			PkName: "id",
		},
	}).(*MessageMediaInstanceDAO)
}

var SharedMessageMediaInstanceDAO *MessageMediaInstanceDAO

func init() {
	dbs.OnReady(func() {
		SharedMessageMediaInstanceDAO = NewMessageMediaInstanceDAO()
	})
}

// EnableMessageMediaInstance 启用条目
func (this *MessageMediaInstanceDAO) EnableMessageMediaInstance(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageMediaInstanceStateEnabled).
		Update()
	return err
}

// DisableMessageMediaInstance 禁用条目
func (this *MessageMediaInstanceDAO) DisableMessageMediaInstance(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageMediaInstanceStateDisabled).
		Update()
	return err
}

// FindEnabledMessageMediaInstance 查找启用中的条目
func (this *MessageMediaInstanceDAO) FindEnabledMessageMediaInstance(tx *dbs.Tx, instanceId int64, cacheMap maps.Map) (*MessageMediaInstance, error) {
	if cacheMap == nil {
		cacheMap = maps.Map{}
	}
	var cacheKey = this.Table + ":record:" + types.String(instanceId)
	var cache = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*MessageMediaInstance), nil
	}

	result, err := this.Query(tx).
		Pk(instanceId).
		Attr("state", MessageMediaInstanceStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}

	cacheMap[cacheKey] = result

	return result.(*MessageMediaInstance), err
}

// CreateMediaInstance 创建媒介实例
func (this *MessageMediaInstanceDAO) CreateMediaInstance(tx *dbs.Tx, name string, mediaType string, params maps.Map, description string) (int64, error) {
	op := NewMessageMediaInstanceOperator()
	op.Name = name
	op.MediaType = mediaType

	// 参数
	if params == nil {
		params = maps.Map{}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return 0, err
	}
	op.Params = paramsJSON

	op.Description = description

	op.IsOn = true
	op.State = MessageMediaInstanceStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateMediaInstance 修改媒介实例
func (this *MessageMediaInstanceDAO) UpdateMediaInstance(tx *dbs.Tx, instanceId int64, name string, mediaType string, params maps.Map, description string, isOn bool) error {
	if instanceId <= 0 {
		return errors.New("invalid instanceId")
	}

	op := NewMessageMediaInstanceOperator()
	op.Id = instanceId
	op.Name = name
	op.MediaType = mediaType

	// 参数
	if params == nil {
		params = maps.Map{}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}
	op.Params = paramsJSON

	op.Description = description
	op.IsOn = isOn
	return this.Save(tx, op)
}

// CountAllEnabledMediaInstances 计算接收人数量
func (this *MessageMediaInstanceDAO) CountAllEnabledMediaInstances(tx *dbs.Tx, mediaType string, keyword string) (int64, error) {
	query := this.Query(tx)
	if len(mediaType) > 0 {
		query.Attr("mediaType", mediaType)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	return query.
		State(MessageMediaInstanceStateEnabled).
		Where("mediaType IN (SELECT `type` FROM " + SharedMessageMediaDAO.Table + " WHERE state=1)").
		Count()
}

// ListAllEnabledMediaInstances 列出单页接收人
func (this *MessageMediaInstanceDAO) ListAllEnabledMediaInstances(tx *dbs.Tx, mediaType string, keyword string, offset int64, size int64) (result []*MessageMediaInstance, err error) {
	query := this.Query(tx)
	if len(mediaType) > 0 {
		query.Attr("mediaType", mediaType)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	_, err = query.
		State(MessageMediaInstanceStateEnabled).
		Where("mediaType IN (SELECT `type` FROM " + SharedMessageMediaDAO.Table + " WHERE state=1)").
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}
