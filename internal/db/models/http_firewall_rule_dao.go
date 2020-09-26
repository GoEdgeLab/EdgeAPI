package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	HTTPFirewallRuleStateEnabled  = 1 // 已启用
	HTTPFirewallRuleStateDisabled = 0 // 已禁用
)

type HTTPFirewallRuleDAO dbs.DAO

func NewHTTPFirewallRuleDAO() *HTTPFirewallRuleDAO {
	return dbs.NewDAO(&HTTPFirewallRuleDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFirewallRules",
			Model:  new(HTTPFirewallRule),
			PkName: "id",
		},
	}).(*HTTPFirewallRuleDAO)
}

var SharedHTTPFirewallRuleDAO = NewHTTPFirewallRuleDAO()

// 初始化
func (this *HTTPFirewallRuleDAO) Init() {
	this.DAOObject.Init()
	this.DAOObject.OnUpdate(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
	this.DAOObject.OnInsert(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
	this.DAOObject.OnDelete(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
}

// 启用条目
func (this *HTTPFirewallRuleDAO) EnableHTTPFirewallRule(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallRuleDAO) DisableHTTPFirewallRule(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPFirewallRuleDAO) FindEnabledHTTPFirewallRule(id uint32) (*HTTPFirewallRule, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPFirewallRuleStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallRule), err
}
