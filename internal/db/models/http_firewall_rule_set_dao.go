package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	HTTPFirewallRuleSetStateEnabled  = 1 // 已启用
	HTTPFirewallRuleSetStateDisabled = 0 // 已禁用
)

type HTTPFirewallRuleSetDAO dbs.DAO

func NewHTTPFirewallRuleSetDAO() *HTTPFirewallRuleSetDAO {
	return dbs.NewDAO(&HTTPFirewallRuleSetDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFirewallRuleSets",
			Model:  new(HTTPFirewallRuleSet),
			PkName: "id",
		},
	}).(*HTTPFirewallRuleSetDAO)
}

var SharedHTTPFirewallRuleSetDAO = NewHTTPFirewallRuleSetDAO()

// 启用条目
func (this *HTTPFirewallRuleSetDAO) EnableHTTPFirewallRuleSet(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleSetStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallRuleSetDAO) DisableHTTPFirewallRuleSet(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleSetStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPFirewallRuleSetDAO) FindEnabledHTTPFirewallRuleSet(id uint32) (*HTTPFirewallRuleSet, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPFirewallRuleSetStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallRuleSet), err
}

// 根据主键查找名称
func (this *HTTPFirewallRuleSetDAO) FindHTTPFirewallRuleSetName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
