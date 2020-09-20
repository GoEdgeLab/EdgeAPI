package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
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

var SharedHTTPAccessLogPolicyDAO = NewHTTPAccessLogPolicyDAO()

// 启用条目
func (this *HTTPAccessLogPolicyDAO) EnableHTTPAccessLogPolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPAccessLogPolicyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPAccessLogPolicyDAO) DisableHTTPAccessLogPolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPAccessLogPolicyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPAccessLogPolicyDAO) FindEnabledHTTPAccessLogPolicy(id int64) (*HTTPAccessLogPolicy, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPAccessLogPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPAccessLogPolicy), err
}

// 根据主键查找名称
func (this *HTTPAccessLogPolicyDAO) FindHTTPAccessLogPolicyName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 查找所有可用策略信息
func (this *HTTPAccessLogPolicyDAO) FindAllEnabledAccessLogPolicies() (result []*HTTPAccessLogPolicy, err error) {
	_, err = this.Query().
		State(HTTPAccessLogPolicyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 组合配置
func (this *HTTPAccessLogPolicyDAO) ComposeAccessLogPolicyConfig(policyId int64) (*serverconfigs.HTTPAccessLogStoragePolicy, error) {
	policy, err := this.FindEnabledHTTPAccessLogPolicy(policyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPAccessLogStoragePolicy{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn == 1
	config.Name = policy.Name
	config.Type = policy.Type

	// 选项
	if IsNotNull(policy.Options) {
		m := map[string]interface{}{}
		err = json.Unmarshal([]byte(policy.Options), &m)
		if err != nil {
			return nil, err
		}
		config.Options = m
	}

	// 条件
	if IsNotNull(policy.Conds) {
		// TODO 需要用更全面的条件管理器来代替RequestCond
		conds := []*shared.RequestCond{}
		err = json.Unmarshal([]byte(policy.Conds), &conds)
		if err != nil {
			return nil, err
		}
		config.Conds = conds
	}

	return config, nil
}