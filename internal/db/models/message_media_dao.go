package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
)

const (
	MessageMediaStateEnabled  = 1 // 已启用
	MessageMediaStateDisabled = 0 // 已禁用
)

type MessageMediaDAO dbs.DAO

func NewMessageMediaDAO() *MessageMediaDAO {
	return dbs.NewDAO(&MessageMediaDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMessageMedias",
			Model:  new(MessageMedia),
			PkName: "id",
		},
	}).(*MessageMediaDAO)
}

var SharedMessageMediaDAO *MessageMediaDAO

func init() {
	dbs.OnReady(func() {
		SharedMessageMediaDAO = NewMessageMediaDAO()
	})
}

// EnableMessageMedia 启用条目
func (this *MessageMediaDAO) EnableMessageMedia(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageMediaStateEnabled).
		Update()
	return err
}

// DisableMessageMedia 禁用条目
func (this *MessageMediaDAO) DisableMessageMedia(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageMediaStateDisabled).
		Update()
	return err
}

// FindEnabledMessageMedia 查找启用中的条目
func (this *MessageMediaDAO) FindEnabledMessageMedia(tx *dbs.Tx, id int64) (*MessageMedia, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", MessageMediaStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MessageMedia), err
}

// 根据主键查找名称
func (this *MessageMediaDAO) FindMessageMediaName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 查询所有可用媒介
func (this *MessageMediaDAO) FindAllEnabledMessageMedias(tx *dbs.Tx) (result []*MessageMedia, err error) {
	_, err = this.Query(tx).
		State(MessageMediaStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// UpdateMessageMedias 设置当前所有可用的媒介
func (this *MessageMediaDAO) UpdateMessageMedias(tx *dbs.Tx, mediaMaps []maps.Map) error {
	// 新的媒介信息
	var mediaTypes = []string{}
	for index, m := range mediaMaps {
		var order = len(mediaMaps) - index
		var mediaType = m.GetString("code")
		mediaTypes = append(mediaTypes, mediaType)

		var name = m.GetString("name")
		var description = m.GetString("description")
		var userDescription = m.GetString("user")
		var isOn = m.GetBool("isOn")

		mediaId, err := this.Query(tx).
			ResultPk().
			Attr("type", mediaType).
			FindInt64Col(0)
		if err != nil {
			return err
		}
		var op = NewMessageMediaOperator()
		if mediaId > 0 {
			op.Id = mediaId
		}
		op.Name = name
		op.Type = mediaType
		op.Description = description
		op.UserDescription = userDescription
		op.IsOn = isOn
		op.Order = order
		op.State = MessageMediaStateEnabled
		err = this.Save(tx, op)
		if err != nil {
			return err
		}
	}

	// 老的媒介信息
	ones, err := this.Query(tx).
		FindAll()
	if err != nil {
		return err
	}
	for _, one := range ones {
		var mediaType = one.(*MessageMedia).Type
		if !lists.ContainsString(mediaTypes, mediaType) {
			err := this.Query(tx).
				Pk(one.(*MessageMedia).Id).
				Set("state", MessageMediaStateDisabled).
				UpdateQuickly()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FindEnabledMediaWithType 根据类型查找媒介
func (this *MessageMediaDAO) FindEnabledMediaWithType(tx *dbs.Tx, mediaType string) (*MessageMedia, error) {
	one, err := this.Query(tx).
		Attr("type", mediaType).
		State(MessageMediaStateEnabled).
		Find()
	if one == nil || err != nil {
		return nil, err
	}
	return one.(*MessageMedia), nil
}
