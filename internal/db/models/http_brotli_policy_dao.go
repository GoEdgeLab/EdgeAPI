package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPBrotliPolicyStateEnabled  = 1 // 已启用
	HTTPBrotliPolicyStateDisabled = 0 // 已禁用
)

type HTTPBrotliPolicyDAO dbs.DAO

func NewHTTPBrotliPolicyDAO() *HTTPBrotliPolicyDAO {
	return dbs.NewDAO(&HTTPBrotliPolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPBrotliPolicies",
			Model:  new(HTTPBrotliPolicy),
			PkName: "id",
		},
	}).(*HTTPBrotliPolicyDAO)
}

var SharedHTTPBrotliPolicyDAO *HTTPBrotliPolicyDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPBrotliPolicyDAO = NewHTTPBrotliPolicyDAO()
	})
}

// EnableHTTPBrotliPolicy 启用条目
func (this *HTTPBrotliPolicyDAO) EnableHTTPBrotliPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPBrotliPolicyStateEnabled).
		Update()
	return err
}

// DisableHTTPBrotliPolicy 禁用条目
func (this *HTTPBrotliPolicyDAO) DisableHTTPBrotliPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPBrotliPolicyStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPBrotliPolicy 查找启用中的条目
func (this *HTTPBrotliPolicyDAO) FindEnabledHTTPBrotliPolicy(tx *dbs.Tx, id int64) (*HTTPBrotliPolicy, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPBrotliPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPBrotliPolicy), err
}

// ComposeBrotliConfig 组合配置
func (this *HTTPBrotliPolicyDAO) ComposeBrotliConfig(tx *dbs.Tx, policyId int64) (*serverconfigs.HTTPBrotliCompressionConfig, error) {
	policy, err := this.FindEnabledHTTPBrotliPolicy(tx, policyId)
	if err != nil {
		return nil, err
	}

	if policy == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPBrotliCompressionConfig{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn
	if IsNotNull(policy.MinLength) {
		minLengthConfig := &shared.SizeCapacity{}
		err = json.Unmarshal(policy.MinLength, minLengthConfig)
		if err != nil {
			return nil, err
		}
		config.MinLength = minLengthConfig
	}
	if IsNotNull(policy.MaxLength) {
		maxLengthConfig := &shared.SizeCapacity{}
		err = json.Unmarshal(policy.MaxLength, maxLengthConfig)
		if err != nil {
			return nil, err
		}
		config.MaxLength = maxLengthConfig
	}
	config.Level = types.Int8(policy.Level)

	if IsNotNull(policy.Conds) {
		condsConfig := &shared.HTTPRequestCondsConfig{}
		err = json.Unmarshal(policy.Conds, condsConfig)
		if err != nil {
			return nil, err
		}
		config.Conds = condsConfig
	}

	return config, nil
}

// CreatePolicy 创建策略
func (this *HTTPBrotliPolicyDAO) CreatePolicy(tx *dbs.Tx, level int, minLengthJSON []byte, maxLengthJSON []byte, condsJSON []byte) (int64, error) {
	op := NewHTTPBrotliPolicyOperator()
	op.State = HTTPBrotliPolicyStateEnabled
	op.IsOn = true
	op.Level = level
	if len(minLengthJSON) > 0 {
		op.MinLength = JSONBytes(minLengthJSON)
	}
	if len(maxLengthJSON) > 0 {
		op.MaxLength = JSONBytes(maxLengthJSON)
	}
	if len(condsJSON) > 0 {
		op.Conds = JSONBytes(condsJSON)
	}
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdatePolicy 修改Policy
func (this *HTTPBrotliPolicyDAO) UpdatePolicy(tx *dbs.Tx, policyId int64, level int, minLengthJSON []byte, maxLengthJSON []byte, condsJSON []byte) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}
	op := NewHTTPBrotliPolicyOperator()
	op.Id = policyId
	op.Level = level
	if len(minLengthJSON) > 0 {
		op.MinLength = JSONBytes(minLengthJSON)
	}
	if len(maxLengthJSON) > 0 {
		op.MaxLength = JSONBytes(maxLengthJSON)
	}
	if len(condsJSON) > 0 {
		op.Conds = JSONBytes(condsJSON)
	}
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// NotifyUpdate 通知更新
func (this *HTTPBrotliPolicyDAO) NotifyUpdate(tx *dbs.Tx, policyId int64) error {
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithBrotliPolicyId(tx, policyId)
	if err != nil {
		return err
	}
	if webId > 0 {
		return SharedHTTPWebDAO.NotifyUpdate(tx, webId)
	}
	return nil
}
