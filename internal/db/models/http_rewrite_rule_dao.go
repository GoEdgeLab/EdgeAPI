package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
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

// Init 初始化
func (this *HTTPRewriteRuleDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPRewriteRule 启用条目
func (this *HTTPRewriteRuleDAO) EnableHTTPRewriteRule(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPRewriteRuleStateEnabled).
		Update()
	return err
}

// DisableHTTPRewriteRule 禁用条目
func (this *HTTPRewriteRuleDAO) DisableHTTPRewriteRule(tx *dbs.Tx, rewriteRuleId int64) error {
	_, err := this.Query(tx).
		Pk(rewriteRuleId).
		Set("state", HTTPRewriteRuleStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, rewriteRuleId)
}

// FindEnabledHTTPRewriteRule 查找启用中的条目
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

// ComposeRewriteRule 构造配置
func (this *HTTPRewriteRuleDAO) ComposeRewriteRule(tx *dbs.Tx, rewriteRuleId int64, cacheMap *utils.CacheMap) (*serverconfigs.HTTPRewriteRule, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":config:" + types.String(rewriteRuleId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*serverconfigs.HTTPRewriteRule), nil
	}

	rule, err := this.FindEnabledHTTPRewriteRule(tx, rewriteRuleId)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPRewriteRule{}
	config.Id = int64(rule.Id)
	config.IsOn = rule.IsOn
	config.Pattern = rule.Pattern
	config.Replace = rule.Replace
	config.Mode = rule.Mode
	config.RedirectStatus = types.Int(rule.RedirectStatus)
	config.ProxyHost = rule.ProxyHost
	config.IsBreak = rule.IsBreak == 1
	config.WithQuery = rule.WithQuery == 1

	// conds
	if len(rule.Conds) > 0 {
		conds := &shared.HTTPRequestCondsConfig{}
		err = json.Unmarshal(rule.Conds, conds)
		if err != nil {
			return nil, err
		}
		config.Conds = conds
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// CreateRewriteRule 创建规则
func (this *HTTPRewriteRuleDAO) CreateRewriteRule(tx *dbs.Tx, pattern string, replace string, mode string, redirectStatus int, isBreak bool, proxyHost string, withQuery bool, isOn bool, condsJSON []byte) (int64, error) {
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

	if len(condsJSON) > 0 {
		op.Conds = condsJSON
	}

	err := this.Save(tx, op)
	return types.Int64(op.Id), err
}

// UpdateRewriteRule 修改规则
func (this *HTTPRewriteRuleDAO) UpdateRewriteRule(tx *dbs.Tx, rewriteRuleId int64, pattern string, replace string, mode string, redirectStatus int, isBreak bool, proxyHost string, withQuery bool, isOn bool, condsJSON []byte) error {
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

	if len(condsJSON) > 0 {
		op.Conds = condsJSON
	}

	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, rewriteRuleId)
}

// NotifyUpdate 通知更新
func (this *HTTPRewriteRuleDAO) NotifyUpdate(tx *dbs.Tx, rewriteRuleId int64) error {
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithRewriteRuleId(tx, rewriteRuleId)
	if err != nil {
		return err
	}
	if webId > 0 {
		return SharedHTTPWebDAO.NotifyUpdate(tx, webId)
	}
	return nil
}
