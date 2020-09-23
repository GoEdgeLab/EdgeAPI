package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	HTTPFirewallRuleGroupStateEnabled  = 1 // 已启用
	HTTPFirewallRuleGroupStateDisabled = 0 // 已禁用
)

type HTTPFirewallRuleGroupDAO dbs.DAO

func NewHTTPFirewallRuleGroupDAO() *HTTPFirewallRuleGroupDAO {
	return dbs.NewDAO(&HTTPFirewallRuleGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFirewallRuleGroups",
			Model:  new(HTTPFirewallRuleGroup),
			PkName: "id",
		},
	}).(*HTTPFirewallRuleGroupDAO)
}

var SharedHTTPFirewallRuleGroupDAO = NewHTTPFirewallRuleGroupDAO()

// 启用条目
func (this *HTTPFirewallRuleGroupDAO) EnableHTTPFirewallRuleGroup(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleGroupStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallRuleGroupDAO) DisableHTTPFirewallRuleGroup(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleGroupStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPFirewallRuleGroupDAO) FindEnabledHTTPFirewallRuleGroup(id uint32) (*HTTPFirewallRuleGroup, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPFirewallRuleGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallRuleGroup), err
}

// 根据主键查找名称
func (this *HTTPFirewallRuleGroupDAO) FindHTTPFirewallRuleGroupName(id uint32) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}
