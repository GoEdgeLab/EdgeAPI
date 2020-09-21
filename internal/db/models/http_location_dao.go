package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	HTTPLocationStateEnabled  = 1 // 已启用
	HTTPLocationStateDisabled = 0 // 已禁用
)

type HTTPLocationDAO dbs.DAO

func NewHTTPLocationDAO() *HTTPLocationDAO {
	return dbs.NewDAO(&HTTPLocationDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPLocations",
			Model:  new(HTTPLocation),
			PkName: "id",
		},
	}).(*HTTPLocationDAO)
}

var SharedHTTPLocationDAO = NewHTTPLocationDAO()

// 启用条目
func (this *HTTPLocationDAO) EnableHTTPLocation(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPLocationStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPLocationDAO) DisableHTTPLocation(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPLocationStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPLocationDAO) FindEnabledHTTPLocation(id uint32) (*HTTPLocation, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPLocationStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPLocation), err
}

// 根据主键查找名称
func (this *HTTPLocationDAO) FindHTTPLocationName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
