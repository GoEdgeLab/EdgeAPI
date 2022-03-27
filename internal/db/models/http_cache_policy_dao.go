package models

import (
	"encoding/json"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
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
func (this *HTTPCachePolicyDAO) CreateCachePolicy(tx *dbs.Tx, isOn bool, name string, description string, capacityJSON []byte, maxKeys int64, maxSizeJSON []byte, storageType string, storageOptionsJSON []byte, syncCompressionCache bool) (int64, error) {
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
	op.SyncCompressionCache = syncCompressionCache

	// 默认的缓存条件
	cacheRef := &serverconfigs.HTTPCacheRef{
		IsOn:                  true,
		Key:                   "${scheme}://${host}${requestURI}",
		Life:                  &shared.TimeDuration{Count: 2, Unit: shared.TimeDurationUnitHour},
		Status:                []int{200},
		MaxSize:               &shared.SizeCapacity{Count: 32, Unit: shared.SizeCapacityUnitMB},
		MinSize:               &shared.SizeCapacity{Count: 0, Unit: shared.SizeCapacityUnitKB},
		SkipResponseSetCookie: true,
		AllowChunkedEncoding:  true,
		Conds: &shared.HTTPRequestCondsConfig{
			IsOn:      true,
			Connector: "or",
			Groups: []*shared.HTTPRequestCondGroup{
				{
					IsOn:      true,
					Connector: "or",
					Conds: []*shared.HTTPRequestCond{
						{
							Type:      "url-extension",
							IsRequest: true,
							Param:     "${requestPathExtension}",
							Operator:  shared.RequestCondOperatorIn,
							Value:     `[".html", ".js", ".css", ".gif", ".png", ".bmp", ".jpeg", ".jpg", ".webp", ".ico", ".pdf", ".ttf", ".eot", ".tiff", ".svg", ".svgz", ".eps", ".woff", ".otf", ".woff2", ".tif", ".csv", ".xls", ".xlsx", ".doc", ".docx", ".ppt", ".pptx", ".wav", ".mp3", ".mp4", ".ogg", ".mid", ".midi"]`,
						},
					},
					Description: "初始化规则",
				},
			},
		},
	}
	refsJSON, err := json.Marshal([]*serverconfigs.HTTPCacheRef{cacheRef})
	if err != nil {
		return 0, err
	}
	op.Refs = refsJSON

	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// CreateDefaultCachePolicy 创建默认的缓存策略
func (this *HTTPCachePolicyDAO) CreateDefaultCachePolicy(tx *dbs.Tx, name string) (int64, error) {
	var capacity = &shared.SizeCapacity{
		Count: 64,
		Unit:  shared.SizeCapacityUnitGB,
	}
	capacityJSON, err := capacity.AsJSON()
	if err != nil {
		return 0, err
	}

	var maxSize = &shared.SizeCapacity{
		Count: 256,
		Unit:  shared.SizeCapacityUnitMB,
	}
	if err != nil {
		return 0, err
	}
	maxSizeJSON, err := maxSize.AsJSON()
	if err != nil {
		return 0, err
	}

	var storageOptions = &serverconfigs.HTTPFileCacheStorage{
		Dir: "/opt/cache",
		MemoryPolicy: &serverconfigs.HTTPCachePolicy{
			Capacity: &shared.SizeCapacity{
				Count: 1,
				Unit:  shared.SizeCapacityUnitGB,
			},
		},
	}
	storageOptionsJSON, err := json.Marshal(storageOptions)
	if err != nil {
		return 0, err
	}

	policyId, err := this.CreateCachePolicy(tx, true, "\""+name+"\"缓存策略", "默认创建的缓存策略", capacityJSON, 0, maxSizeJSON, serverconfigs.CachePolicyStorageFile, storageOptionsJSON, false)
	if err != nil {
		return 0, err
	}
	return policyId, nil
}

// UpdateCachePolicy 修改缓存策略
func (this *HTTPCachePolicyDAO) UpdateCachePolicy(tx *dbs.Tx, policyId int64, isOn bool, name string, description string, capacityJSON []byte, maxKeys int64, maxSizeJSON []byte, storageType string, storageOptionsJSON []byte, syncCompressionCache bool) error {
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
	op.SyncCompressionCache = syncCompressionCache
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// ComposeCachePolicy 组合配置
func (this *HTTPCachePolicyDAO) ComposeCachePolicy(tx *dbs.Tx, policyId int64, cacheMap *utils.CacheMap) (*serverconfigs.HTTPCachePolicy, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":config:" + types.String(policyId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*serverconfigs.HTTPCachePolicy), nil
	}

	policy, err := this.FindEnabledHTTPCachePolicy(tx, policyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	config := &serverconfigs.HTTPCachePolicy{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn
	config.Name = policy.Name
	config.Description = policy.Description
	config.SyncCompressionCache = policy.SyncCompressionCache == 1

	// capacity
	if IsNotNull(policy.Capacity) {
		capacityConfig := &shared.SizeCapacity{}
		err = json.Unmarshal(policy.Capacity, capacityConfig)
		if err != nil {
			return nil, err
		}
		config.Capacity = capacityConfig
	}

	config.MaxKeys = types.Int64(policy.MaxKeys)

	// max size
	if IsNotNull(policy.MaxSize) {
		maxSizeConfig := &shared.SizeCapacity{}
		err = json.Unmarshal(policy.MaxSize, maxSizeConfig)
		if err != nil {
			return nil, err
		}
		config.MaxSize = maxSizeConfig
	}

	config.Type = policy.Type

	// options
	if IsNotNull(policy.Options) {
		m := map[string]interface{}{}
		err = json.Unmarshal(policy.Options, &m)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		config.Options = m
	}

	// refs
	if IsNotNull(policy.Refs) {
		refs := []*serverconfigs.HTTPCacheRef{}
		err = json.Unmarshal(policy.Refs, &refs)
		if err != nil {
			return nil, err
		}
		config.CacheRefs = refs
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// CountAllEnabledHTTPCachePolicies 计算可用缓存策略数量
func (this *HTTPCachePolicyDAO) CountAllEnabledHTTPCachePolicies(tx *dbs.Tx, clusterId int64, keyword string, storageType string) (int64, error) {
	query := this.Query(tx).
		State(HTTPCachePolicyStateEnabled)
	if clusterId > 0 {
		query.Where("id IN (SELECT cachePolicyId FROM " + SharedNodeClusterDAO.Table + " WHERE id=:clusterId)")
		query.Param("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if len(storageType) > 0 {
		query.Attr("type", storageType)
	}
	return query.Count()
}

// ListEnabledHTTPCachePolicies 列出单页的缓存策略
func (this *HTTPCachePolicyDAO) ListEnabledHTTPCachePolicies(tx *dbs.Tx, clusterId int64, keyword string, storageType string, offset int64, size int64) ([]*serverconfigs.HTTPCachePolicy, error) {
	query := this.Query(tx).
		State(HTTPCachePolicyStateEnabled)
	if clusterId > 0 {
		query.Where("id IN (SELECT cachePolicyId FROM " + SharedNodeClusterDAO.Table + " WHERE id=:clusterId)")
		query.Param("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if len(storageType) > 0 {
		query.Attr("type", storageType)
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
		cachePolicyConfig, err := this.ComposeCachePolicy(tx, policyId, nil)
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
