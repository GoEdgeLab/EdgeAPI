package models

import (
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPRewriteRuleStateEnabled  = 1 // 已启用
	HTTPRewriteRuleStateDisabled = 0 // 已禁用
)

type HTTPRewriteRuleDAO dbs.DAO

func NewHTTPRewriteRuleDAO() *HTTPRewriteRuleDAO {
	return dbs.NewDAO(&HTTPRewriteRuleDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPRewriteRules",
			Model:  new(HTTPRewriteRule),
			PkName: "id",
		},
	}).(*HTTPRewriteRuleDAO)
}

var SharedHTTPRewriteRuleDAO *HTTPRewriteRuleDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPRewriteRuleDAO = NewHTTPRewriteRuleDAO()
	})
}

// 初始化
func (this *HTTPRewriteRuleDAO) Init() {
	this.DAOObject.Init()
	this.DAOObject.OnUpdate(func() error {
		return SharedSysEventDAO.CreateEvent(nil, NewServerChangeEvent())
	})
	this.DAOObject.OnInsert(func() error {
		return SharedSysEventDAO.CreateEvent(nil, NewServerChangeEvent())
	})
	this.DAOObject.OnDelete(func() error {
		return SharedSysEventDAO.CreateEvent(nil, NewServerChangeEvent())
	})
}

// 启用条目
func (this *HTTPRewriteRuleDAO) EnableHTTPRewriteRule(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPRewriteRuleStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPRewriteRuleDAO) DisableHTTPRewriteRule(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPRewriteRuleStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPRewriteRuleDAO) FindEnabledHTTPRewriteRule(tx *dbs.Tx, id int64) (*HTTPRewriteRule, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPRewriteRuleStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPRewriteRule), err
}

// 构造配置
func (this *HTTPRewriteRuleDAO) ComposeRewriteRule(tx *dbs.Tx, rewriteRuleId int64) (*serverconfigs.HTTPRewriteRule, error) {
	rule, err := this.FindEnabledHTTPRewriteRule(tx, rewriteRuleId)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPRewriteRule{}
	config.Id = int64(rule.Id)
	config.IsOn = rule.IsOn == 1
	config.Pattern = rule.Pattern
	config.Replace = rule.Replace
	config.Mode = rule.Mode
	config.RedirectStatus = types.Int(rule.RedirectStatus)
	config.ProxyHost = rule.ProxyHost
	config.IsBreak = rule.IsBreak == 1
	config.WithQuery = rule.WithQuery == 1
	return config, nil
}

// 创建规则
func (this *HTTPRewriteRuleDAO) CreateRewriteRule(tx *dbs.Tx, pattern string, replace string, mode string, redirectStatus int, isBreak bool, proxyHost string, withQuery bool, isOn bool) (int64, error) {
	op := NewHTTPRewriteRuleOperator()
	op.State = HTTPRewriteRuleStateEnabled
	op.IsOn = isOn

	op.Pattern = pattern
	op.Replace = replace
	op.Mode = mode
	op.RedirectStatus = redirectStatus
	op.IsBreak = isBreak
	op.WithQuery = withQuery
	op.ProxyHost = proxyHost
	err := this.Save(tx, op)
	return types.Int64(op.Id), err
}

// 修改规则
func (this *HTTPRewriteRuleDAO) UpdateRewriteRule(tx *dbs.Tx, rewriteRuleId int64, pattern string, replace string, mode string, redirectStatus int, isBreak bool, proxyHost string, withQuery bool, isOn bool) error {
	if rewriteRuleId <= 0 {
		return errors.New("invalid rewriteRuleId")
	}
	op := NewHTTPRewriteRuleOperator()
	op.Id = rewriteRuleId
	op.IsOn = isOn
	op.Pattern = pattern
	op.Replace = replace
	op.Mode = mode
	op.RedirectStatus = redirectStatus
	op.IsBreak = isBreak
	op.WithQuery = withQuery
	op.ProxyHost = proxyHost
	err := this.Save(tx, op)
	return err
}
