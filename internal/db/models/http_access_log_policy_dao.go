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

var SharedHTTPAccessLogPolicyDAO *HTTPAccessLogPolicyDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPAccessLogPolicyDAO = NewHTTPAccessLogPolicyDAO()
	})
}

// 初始化
func (this *HTTPAccessLogPolicyDAO) Init() {
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
func (this *HTTPAccessLogPolicyDAO) EnableHTTPAccessLogPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPAccessLogPolicyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPAccessLogPolicyDAO) DisableHTTPAccessLogPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPAccessLogPolicyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
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

// 根据主键查找名称
func (this *HTTPAccessLogPolicyDAO) FindHTTPAccessLogPolicyName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 查找所有可用策略信息
func (this *HTTPAccessLogPolicyDAO) FindAllEnabledAccessLogPolicies(tx *dbs.Tx) (result []*HTTPAccessLogPolicy, err error) {
	_, err = this.Query(tx).
		State(HTTPAccessLogPolicyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 组合配置
func (this *HTTPAccessLogPolicyDAO) ComposeAccessLogPolicyConfig(tx *dbs.Tx, policyId int64) (*serverconfigs.HTTPAccessLogStoragePolicy, error) {
	policy, err := this.FindEnabledHTTPAccessLogPolicy(tx, policyId)
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
		condsConfig := &shared.HTTPRequestCondsConfig{}
		err = json.Unmarshal([]byte(policy.Conds), condsConfig)
		if err != nil {
			return nil, err
		}
		config.Conds = condsConfig
	}

	return config, nil
}
