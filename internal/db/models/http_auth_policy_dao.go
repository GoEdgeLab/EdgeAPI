package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPAuthPolicyStateEnabled  = 1 // 已启用
	HTTPAuthPolicyStateDisabled = 0 // 已禁用
)

type HTTPAuthPolicyDAO dbs.DAO

func NewHTTPAuthPolicyDAO() *HTTPAuthPolicyDAO {
	return dbs.NewDAO(&HTTPAuthPolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPAuthPolicies",
			Model:  new(HTTPAuthPolicy),
			PkName: "id",
		},
	}).(*HTTPAuthPolicyDAO)
}

var SharedHTTPAuthPolicyDAO *HTTPAuthPolicyDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPAuthPolicyDAO = NewHTTPAuthPolicyDAO()
	})
}

// EnableHTTPAuthPolicy 启用条目
func (this *HTTPAuthPolicyDAO) EnableHTTPAuthPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPAuthPolicyStateEnabled).
		Update()
	return err
}

// DisableHTTPAuthPolicy 禁用条目
func (this *HTTPAuthPolicyDAO) DisableHTTPAuthPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPAuthPolicyStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPAuthPolicy 查找启用中的条目
func (this *HTTPAuthPolicyDAO) FindEnabledHTTPAuthPolicy(tx *dbs.Tx, id int64) (*HTTPAuthPolicy, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPAuthPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPAuthPolicy), err
}

// CreateHTTPAuthPolicy 创建策略
func (this *HTTPAuthPolicyDAO) CreateHTTPAuthPolicy(tx *dbs.Tx, name string, methodType string, paramsJSON []byte) (int64, error) {
	op := NewHTTPAuthPolicyOperator()
	op.Name = name
	op.Type = methodType
	op.Params = paramsJSON
	op.IsOn = true
	op.State = HTTPAuthPolicyStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateHTTPAuthPolicy 修改策略
func (this *HTTPAuthPolicyDAO) UpdateHTTPAuthPolicy(tx *dbs.Tx, policyId int64, name string, paramsJSON []byte, isOn bool) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}
	op := NewHTTPAuthPolicyOperator()
	op.Id = policyId
	op.Name = name
	op.Params = paramsJSON
	op.IsOn = isOn
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// ComposePolicyConfig 组合配置
func (this *HTTPAuthPolicyDAO) ComposePolicyConfig(tx *dbs.Tx, policyId int64, cacheMap *utils.CacheMap) (*serverconfigs.HTTPAuthPolicy, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":config:" + types.String(policyId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*serverconfigs.HTTPAuthPolicy), nil
	}

	policy, err := this.FindEnabledHTTPAuthPolicy(tx, policyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	var config = &serverconfigs.HTTPAuthPolicy{
		Id:   int64(policy.Id),
		Name: policy.Name,
		IsOn: policy.IsOn == 1,
		Type: policy.Type,
	}

	var params map[string]interface{}
	if IsNotNull(policy.Params) {
		err = json.Unmarshal(policy.Params, &params)
		if err != nil {
			return nil, err
		}
		config.Params = params
	}
	config.Params = params

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// NotifyUpdate 通知更改
func (this *HTTPAuthPolicyDAO) NotifyUpdate(tx *dbs.Tx, policyId int64) error {
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithHTTPAuthPolicyId(tx, policyId)
	if err != nil {
		return err
	}
	if webId > 0 {
		return SharedHTTPWebDAO.NotifyUpdate(tx, webId)
	}
	return nil
}
