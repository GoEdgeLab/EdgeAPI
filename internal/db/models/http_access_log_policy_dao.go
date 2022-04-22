package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	HTTPAccessLogPolicyStateEnabled  = 1 // 已启用
	HTTPAccessLogPolicyStateDisabled = 0 // 已禁用
)

type HTTPAccessLogPolicyDAO dbs.DAO

func NewHTTPAccessLogPolicyDAO() *HTTPAccessLogPolicyDAO {
	return dbs.NewDAO(&HTTPAccessLogPolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPAccessLogPolicies",
			Model:  new(HTTPAccessLogPolicy),
			PkName: "id",
		},
	}).(*HTTPAccessLogPolicyDAO)
}

var SharedHTTPAccessLogPolicyDAO *HTTPAccessLogPolicyDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPAccessLogPolicyDAO = NewHTTPAccessLogPolicyDAO()
	})
}

// Init 初始化
func (this *HTTPAccessLogPolicyDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPAccessLogPolicy 启用条目
func (this *HTTPAccessLogPolicyDAO) EnableHTTPAccessLogPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPAccessLogPolicyStateEnabled).
		Update()
	return err
}

// DisableHTTPAccessLogPolicy 禁用条目
func (this *HTTPAccessLogPolicyDAO) DisableHTTPAccessLogPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPAccessLogPolicyStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPAccessLogPolicy 查找启用中的条目
func (this *HTTPAccessLogPolicyDAO) FindEnabledHTTPAccessLogPolicy(tx *dbs.Tx, id int64) (*HTTPAccessLogPolicy, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPAccessLogPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPAccessLogPolicy), err
}

// FindHTTPAccessLogPolicyName 根据主键查找名称
func (this *HTTPAccessLogPolicyDAO) FindHTTPAccessLogPolicyName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CountAllEnabledPolicies 计算策略数量
func (this *HTTPAccessLogPolicyDAO) CountAllEnabledPolicies(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(HTTPAccessLogPolicyStateEnabled).
		Count()
}

// ListEnabledPolicies 查找所有可用策略信息
func (this *HTTPAccessLogPolicyDAO) ListEnabledPolicies(tx *dbs.Tx, offset int64, size int64) (result []*HTTPAccessLogPolicy, err error) {
	_, err = this.Query(tx).
		State(HTTPAccessLogPolicyStateEnabled).
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledAndOnPolicies 获取所有的策略信息
func (this *HTTPAccessLogPolicyDAO) FindAllEnabledAndOnPolicies(tx *dbs.Tx) (result []*HTTPAccessLogPolicy, err error) {
	_, err = this.Query(tx).
		State(HTTPAccessLogPolicyStateEnabled).
		Attr("isOn", true).
		Slice(&result).
		FindAll()
	return
}

// CreatePolicy 创建策略
func (this *HTTPAccessLogPolicyDAO) CreatePolicy(tx *dbs.Tx, name string, policyType string, optionsJSON []byte, condsJSON []byte, isPublic bool, firewallOnly bool) (policyId int64, err error) {
	var op = NewHTTPAccessLogPolicyOperator()
	op.Name = name
	op.Type = policyType
	if len(optionsJSON) > 0 {
		op.Options = optionsJSON
	}
	if len(condsJSON) > 0 {
		op.Conds = condsJSON
	}
	op.IsPublic = isPublic
	op.IsOn = true
	op.FirewallOnly = firewallOnly
	op.State = HTTPAccessLogPolicyStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdatePolicy 修改策略
func (this *HTTPAccessLogPolicyDAO) UpdatePolicy(tx *dbs.Tx, policyId int64, name string, optionsJSON []byte, condsJSON []byte, isPublic bool, firewallOnly bool, isOn bool) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	oldOne, err := this.Query(tx).
		Pk(policyId).
		Find()
	if err != nil {
		return err
	}
	if oldOne == nil {
		return nil
	}

	var op = NewHTTPAccessLogPolicyOperator()
	op.Id = policyId
	op.Name = name
	if len(optionsJSON) > 0 {
		op.Options = optionsJSON
	} else {
		op.Options = "{}"
	}
	if len(condsJSON) > 0 {
		op.Conds = condsJSON
	} else {
		op.Conds = "{}"
	}

	// 版本号总是加1
	op.Version = dbs.SQL("version+1")

	op.IsPublic = isPublic
	op.FirewallOnly = firewallOnly
	op.IsOn = isOn
	return this.Save(tx, op)
}

// CancelAllPublicPolicies 取消别的公用的策略
func (this *HTTPAccessLogPolicyDAO) CancelAllPublicPolicies(tx *dbs.Tx) error {
	return this.Query(tx).
		State(HTTPAccessLogPolicyStateEnabled).
		Set("isPublic", 0).
		UpdateQuickly()
}

// FindCurrentPublicPolicyId 取得当前的公用策略
func (this *HTTPAccessLogPolicyDAO) FindCurrentPublicPolicyId(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(HTTPAccessLogPolicyStateEnabled).
		Attr("isPublic", 1).
		ResultPk().
		FindInt64Col(0)
}
