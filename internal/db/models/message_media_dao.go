package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
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

// FindMessageMediaName 根据主键查找名称
func (this *MessageMediaDAO) FindMessageMediaName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindAllEnabledMessageMedias 查询所有可用媒介
func (this *MessageMediaDAO) FindAllEnabledMessageMedias(tx *dbs.Tx) (result []*MessageMedia, err error) {
	_, err = this.Query(tx).
		State(MessageMediaStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}
