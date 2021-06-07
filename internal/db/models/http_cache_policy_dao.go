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
	HTTPCachePolicyStateEnabled  = 1 // 已启用
	HTTPCachePolicyStateDisabled = 0 // 已禁用
)

type HTTPCachePolicyDAO dbs.DAO

func NewHTTPCachePolicyDAO() *HTTPCachePolicyDAO {
	return dbs.NewDAO(&HTTPCachePolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPCachePolicies",
			Model:  new(HTTPCachePolicy),
			PkName: "id",
		},
	}).(*HTTPCachePolicyDAO)
}

var SharedHTTPCachePolicyDAO *HTTPCachePolicyDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPCachePolicyDAO = NewHTTPCachePolicyDAO()
	})
}

// Init 初始化
func (this *HTTPCachePolicyDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPCachePolicy 启用条目
func (this *HTTPCachePolicyDAO) EnableHTTPCachePolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPCachePolicyStateEnabled).
		Update()
	return err
}

// DisableHTTPCachePolicy 禁用条目
func (this *HTTPCachePolicyDAO) DisableHTTPCachePolicy(tx *dbs.Tx, policyId int64) error {
	_, err := this.Query(tx).
		Pk(policyId).
		Set("state", HTTPCachePolicyStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// FindEnabledHTTPCachePolicy 查找启用中的条目
func (this *HTTPCachePolicyDAO) FindEnabledHTTPCachePolicy(tx *dbs.Tx, id int64) (*HTTPCachePolicy, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPCachePolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPCachePolicy), err
}

// FindHTTPCachePolicyName 根据主键查找名称
func (this *HTTPCachePolicyDAO) FindHTTPCachePolicyName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindAllEnabledCachePolicies 查找所有可用的缓存策略
func (this *HTTPCachePolicyDAO) FindAllEnabledCachePolicies(tx *dbs.Tx) (result []*HTTPCachePolicy, err error) {
	_, err = this.Query(tx).
		State(HTTPCachePolicyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CreateCachePolicy 创建缓存策略
func (this *HTTPCachePolicyDAO) CreateCachePolicy(tx *dbs.Tx, isOn bool, name string, description string, capacityJSON []byte, maxKeys int64, maxSizeJSON []byte, storageType string, storageOptionsJSON []byte) (int64, error) {
	op := NewHTTPCachePolicyOperator()
	op.State = HTTPCachePolicyStateEnabled
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	if len(capacityJSON) > 0 {
		op.Capacity = capacityJSON
	}
	op.MaxKeys = maxKeys
	if len(maxSizeJSON) > 0 {
		op.MaxSize = maxSizeJSON
	}
	op.Type = storageType
	if len(storageOptionsJSON) > 0 {
		op.Options = storageOptionsJSON
	}
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateCachePolicy 修改缓存策略
func (this *HTTPCachePolicyDAO) UpdateCachePolicy(tx *dbs.Tx, policyId int64, isOn bool, name string, description string, capacityJSON []byte, maxKeys int64, maxSizeJSON []byte, storageType string, storageOptionsJSON []byte) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	op := NewHTTPCachePolicyOperator()
	op.Id = policyId
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	if len(capacityJSON) > 0 {
		op.Capacity = capacityJSON
	}
	op.MaxKeys = maxKeys
	if len(maxSizeJSON) > 0 {
		op.MaxSize = maxSizeJSON
	}
	op.Type = storageType
	if len(storageOptionsJSON) > 0 {
		op.Options = storageOptionsJSON
	}
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// ComposeCachePolicy 组合配置
func (this *HTTPCachePolicyDAO) ComposeCachePolicy(tx *dbs.Tx, policyId int64) (*serverconfigs.HTTPCachePolicy, error) {
	policy, err := this.FindEnabledHTTPCachePolicy(tx, policyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	config := &serverconfigs.HTTPCachePolicy{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn == 1
	config.Name = policy.Name
	config.Description = policy.Description

	// capacity
	if IsNotNull(policy.Capacity) {
		capacityConfig := &shared.SizeCapacity{}
		err = json.Unmarshal([]byte(policy.Capacity), capacityConfig)
		if err != nil {
			return nil, err
		}
		config.Capacity = capacityConfig
	}

	config.MaxKeys = types.Int64(policy.MaxKeys)

	// max size
	if IsNotNull(policy.MaxSize) {
		maxSizeConfig := &shared.SizeCapacity{}
		err = json.Unmarshal([]byte(policy.MaxSize), maxSizeConfig)
		if err != nil {
			return nil, err
		}
		config.MaxSize = maxSizeConfig
	}

	config.Type = policy.Type

	// options
	if IsNotNull(policy.Options) {
		m := map[string]interface{}{}
		err = json.Unmarshal([]byte(policy.Options), &m)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		config.Options = m
	}

	// refs
	if IsNotNull(policy.Refs) {
		refs := []*serverconfigs.HTTPCacheRef{}
		err = json.Unmarshal([]byte(policy.Refs), &refs)
		if err != nil {
			return nil, err
		}
		config.CacheRefs = refs
	}

	return config, nil
}

// CountAllEnabledHTTPCachePolicies 计算可用缓存策略数量
func (this *HTTPCachePolicyDAO) CountAllEnabledHTTPCachePolicies(tx *dbs.Tx, keyword string) (int64, error) {
	query := this.Query(tx).
		State(HTTPCachePolicyStateEnabled)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	return query.Count()
}

// ListEnabledHTTPCachePolicies 列出单页的缓存策略
func (this *HTTPCachePolicyDAO) ListEnabledHTTPCachePolicies(tx *dbs.Tx, keyword string, offset int64, size int64) ([]*serverconfigs.HTTPCachePolicy, error) {
	query := this.Query(tx).
		State(HTTPCachePolicyStateEnabled)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	ones, err := query.
		ResultPk().
		Offset(offset).
		Limit(size).
		DescPk().
		FindAll()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	cachePolicyIds := []int64{}
	for _, one := range ones {
		cachePolicyIds = append(cachePolicyIds, int64(one.(*HTTPCachePolicy).Id))
	}
	if len(cachePolicyIds) == 0 {
		return nil, nil
	}

	cachePolicies := []*serverconfigs.HTTPCachePolicy{}
	for _, policyId := range cachePolicyIds {
		cachePolicyConfig, err := this.ComposeCachePolicy(tx, policyId)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		cachePolicies = append(cachePolicies, cachePolicyConfig)
	}
	return cachePolicies, nil
}

// UpdatePolicyRefs 设置默认的缓存条件
func (this *HTTPCachePolicyDAO) UpdatePolicyRefs(tx *dbs.Tx, policyId int64, refsJSON []byte) error {
	if len(refsJSON) == 0 {
		return nil
	}
	err := this.Query(tx).
		Pk(policyId).
		Set("refs", refsJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// NotifyUpdate 通知更新
func (this *HTTPCachePolicyDAO) NotifyUpdate(tx *dbs.Tx, policyId int64) error {
	webIds, err := SharedHTTPWebDAO.FindAllWebIdsWithCachePolicyId(tx, policyId)
	if err != nil {
		return err
	}
	for _, webId := range webIds {
		err := SharedHTTPWebDAO.NotifyUpdate(tx, webId)
		if err != nil {
			return err
		}
	}

	clusterIds, err := SharedNodeClusterDAO.FindAllEnabledNodeClusterIdsWithCachePolicyId(tx, policyId)
	if err != nil {
		return err
	}
	for _, clusterId := range clusterIds {
		err := SharedNodeClusterDAO.NotifyUpdate(tx, clusterId)
		if err != nil {
			return err
		}
	}
	return nil
}
