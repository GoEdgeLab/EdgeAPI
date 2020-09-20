package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	HTTPFirewallPolicyStateEnabled  = 1 // 已启用
	HTTPFirewallPolicyStateDisabled = 0 // 已禁用
)

type HTTPFirewallPolicyDAO dbs.DAO

func NewHTTPFirewallPolicyDAO() *HTTPFirewallPolicyDAO {
	return dbs.NewDAO(&HTTPFirewallPolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFirewallPolicies",
			Model:  new(HTTPFirewallPolicy),
			PkName: "id",
		},
	}).(*HTTPFirewallPolicyDAO)
}

var SharedHTTPFirewallPolicyDAO = NewHTTPFirewallPolicyDAO()

// 启用条目
func (this *HTTPFirewallPolicyDAO) EnableHTTPFirewallPolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallPolicyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallPolicyDAO) DisableHTTPFirewallPolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallPolicyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPFirewallPolicyDAO) FindEnabledHTTPFirewallPolicy(id int64) (*HTTPFirewallPolicy, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPFirewallPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallPolicy), err
}

// 根据主键查找名称
func (this *HTTPFirewallPolicyDAO) FindHTTPFirewallPolicyName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 查找所有可用策略
func (this *HTTPFirewallPolicyDAO) FindAllEnabledFirewallPolicies() (result []*HTTPFirewallPolicy, err error) {
	_, err = this.Query().
		State(HTTPFirewallPolicyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
