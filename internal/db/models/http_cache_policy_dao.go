package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	HTTPCachePolicyStateEnabled  = 1 // 已启用
	HTTPCachePolicyStateDisabled = 0 // 已禁用
)

type HTTPCachePolicyDAO dbs.DAO

func NewHTTPCachePolicyDAO() *HTTPCachePolicyDAO {
	return dbs.NewDAO(&HTTPCachePolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPCachePolicies",
			Model:  new(HTTPCachePolicy),
			PkName: "id",
		},
	}).(*HTTPCachePolicyDAO)
}

var SharedHTTPCachePolicyDAO = NewHTTPCachePolicyDAO()

// 启用条目
func (this *HTTPCachePolicyDAO) EnableHTTPCachePolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPCachePolicyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPCachePolicyDAO) DisableHTTPCachePolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPCachePolicyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPCachePolicyDAO) FindEnabledHTTPCachePolicy(id int64) (*HTTPCachePolicy, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPCachePolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPCachePolicy), err
}

// 根据主键查找名称
func (this *HTTPCachePolicyDAO) FindHTTPCachePolicyName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 查找所有可用的缓存策略
func (this *HTTPCachePolicyDAO) FindAllEnabledCachePolicies() (result []*HTTPCachePolicy, err error) {
	_, err = this.Query().
		State(HTTPCachePolicyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
