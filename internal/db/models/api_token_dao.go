package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ApiTokenStateEnabled  = 1 // 已启用
	ApiTokenStateDisabled = 0 // 已禁用
)

type ApiTokenDAO dbs.DAO

func NewApiTokenDAO() *ApiTokenDAO {
	return dbs.NewDAO(&ApiTokenDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeApiTokens",
			Model:  new(ApiToken),
			PkName: "id",
		},
	}).(*ApiTokenDAO)
}

var SharedApiTokenDAO = NewApiTokenDAO()

// 启用条目
func (this *ApiTokenDAO) EnableApiToken(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ApiTokenStateEnabled).
		Update()
}

// 禁用条目
func (this *ApiTokenDAO) DisableApiToken(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ApiTokenStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *ApiTokenDAO) FindEnabledApiToken(id uint32) (*ApiToken, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ApiTokenStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ApiToken), err
}

// 获取节点Token信息
// TODO 需要添加缓存
func (this *ApiTokenDAO) FindEnabledTokenWithNode(nodeId string) (*ApiToken, error) {
	one, err := this.Query().
		Attr("nodeId", nodeId).
		State(ApiTokenStateEnabled).
		Find()
	if one != nil {
		return one.(*ApiToken), nil
	}
	return nil, err
}
