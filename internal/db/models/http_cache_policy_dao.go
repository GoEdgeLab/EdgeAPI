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

// 初始化
func (this *HTTPCachePolicyDAO) Init() {
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
func (this *HTTPCachePolicyDAO) EnableHTTPCachePolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPCachePolicyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPCachePolicyDAO) DisableHTTPCachePolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPCachePolicyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPCachePolicyDAO) FindEnabledHTTPCachePolicy(id int64) (*HTTPCachePolicy, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPCachePolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPCachePolicy), err
}

// 根据主键查找名称
func (this *HTTPCachePolicyDAO) FindHTTPCachePolicyName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 查找所有可用的缓存策略
func (this *HTTPCachePolicyDAO) FindAllEnabledCachePolicies() (result []*HTTPCachePolicy, err error) {
	_, err = this.Query().
		State(HTTPCachePolicyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 创建缓存策略
func (this *HTTPCachePolicyDAO) CreateCachePolicy(isOn bool, name string, description string, capacityJSON []byte, maxKeys int64, maxSizeJSON []byte, storageType string, storageOptionsJSON []byte) (int64, error) {
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
	err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改缓存策略
func (this *HTTPCachePolicyDAO) UpdateCachePolicy(policyId int64, isOn bool, name string, description string, capacityJSON []byte, maxKeys int64, maxSizeJSON []byte, storageType string, storageOptionsJSON []byte) error {
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
	err := this.Save(op)
	return errors.Wrap(err)
}

// 组合配置
func (this *HTTPCachePolicyDAO) ComposeCachePolicy(policyId int64) (*serverconfigs.HTTPCachePolicy, error) {
	policy, err := this.FindEnabledHTTPCachePolicy(policyId)
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

	return config, nil
}

// 计算可用缓存策略数量
func (this *HTTPCachePolicyDAO) CountAllEnabledHTTPCachePolicies() (int64, error) {
	return this.Query().
		State(HTTPCachePolicyStateEnabled).
		Count()
}

// 列出单页的缓存策略
func (this *HTTPCachePolicyDAO) ListEnabledHTTPCachePolicies(offset int64, size int64) ([]*serverconfigs.HTTPCachePolicy, error) {
	ones, err := this.Query().
		State(HTTPCachePolicyStateEnabled).
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
		cachePolicyConfig, err := this.ComposeCachePolicy(policyId)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		cachePolicies = append(cachePolicies, cachePolicyConfig)
	}
	return cachePolicies, nil
}
