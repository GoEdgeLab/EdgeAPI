package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	HTTPHeaderStateEnabled  = 1 // 已启用
	HTTPHeaderStateDisabled = 0 // 已禁用
)

type HTTPHeaderDAO dbs.DAO

func NewHTTPHeaderDAO() *HTTPHeaderDAO {
	return dbs.NewDAO(&HTTPHeaderDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPHeaders",
			Model:  new(HTTPHeader),
			PkName: "id",
		},
	}).(*HTTPHeaderDAO)
}

var SharedHTTPHeaderDAO = NewHTTPHeaderDAO()

// 启用条目
func (this *HTTPHeaderDAO) EnableHTTPHeader(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPHeaderStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPHeaderDAO) DisableHTTPHeader(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPHeaderStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPHeaderDAO) FindEnabledHTTPHeader(id uint32) (*HTTPHeader, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPHeaderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPHeader), err
}

// 根据主键查找名称
func (this *HTTPHeaderDAO) FindHTTPHeaderName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
