package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"sort"
	"strings"
)

const (
	MetricItemStateEnabled  = 1 // 已启用
	MetricItemStateDisabled = 0 // 已禁用
)

type MetricItemDAO dbs.DAO

func NewMetricItemDAO() *MetricItemDAO {
	return dbs.NewDAO(&MetricItemDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMetricItems",
			Model:  new(MetricItem),
			PkName: "id",
		},
	}).(*MetricItemDAO)
}

var SharedMetricItemDAO *MetricItemDAO

func init() {
	dbs.OnReady(func() {
		SharedMetricItemDAO = NewMetricItemDAO()
	})
}

// EnableMetricItem 启用条目
func (this *MetricItemDAO) EnableMetricItem(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MetricItemStateEnabled).
		Update()
	return err
}

// DisableMetricItem 禁用条目
func (this *MetricItemDAO) DisableMetricItem(tx *dbs.Tx, itemId int64) error {
	isPublic, err := this.Query(tx).
		Pk(itemId).
		Result("isPublic").
		FindIntCol(0)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(itemId).
		Set("state", MetricItemStateDisabled).
		Update()
	if err != nil {
		return err
	}

	// 通知更新
	err = this.NotifyUpdate(tx, itemId, isPublic == 1)
	if err != nil {
		return err
	}

	// 删除统计数据
	err = SharedMetricStatDAO.DeleteItemStats(tx, itemId)
	if err != nil {
		return err
	}
	return nil
}

// FindEnabledMetricItem 查找启用中的条目
func (this *MetricItemDAO) FindEnabledMetricItem(tx *dbs.Tx, id int64) (*MetricItem, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", MetricItemStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MetricItem), err
}

// FindMetricItemName 根据主键查找名称
func (this *MetricItemDAO) FindMetricItemName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateItem 创建指标
func (this *MetricItemDAO) CreateItem(tx *dbs.Tx, code string, category string, name string, keys []string, period int32, periodUnit string, value string, isPublic bool) (int64, error) {
	sort.Strings(keys)

	op := NewMetricItemOperator()
	op.Code = code
	op.Category = category
	op.Name = name
	if len(keys) > 0 {
		keysJSON, err := json.Marshal(keys)
		if err != nil {
			return 0, err
		}
		op.Keys = keysJSON
	} else {
		op.Keys = "[]"
	}
	op.Period = period
	op.PeriodUnit = periodUnit
	op.Value = value
	op.IsPublic = isPublic
	op.IsOn = true
	op.State = MetricItemStateEnabled
	itemId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	if isPublic {
		err = this.NotifyUpdate(tx, itemId, isPublic)
		if err != nil {
			return 0, err
		}
	}

	return itemId, nil
}

// UpdateItem 修改\指标
func (this *MetricItemDAO) UpdateItem(tx *dbs.Tx, itemId int64, name string, keys []string, period int32, periodUnit string, value string, isOn bool, isPublic bool) error {
	if itemId <= 0 {
		return errors.New("invalid itemId")
	}

	sort.Strings(keys)

	// 是否有变化
	oldItem, err := this.FindEnabledMetricItem(tx, itemId)
	if err != nil {
		return err
	}
	if oldItem == nil {
		return nil
	}
	oldIsPublic := oldItem.IsPublic == 1
	var versionChanged = false
	if strings.Join(oldItem.DecodeKeys(), "&") != strings.Join(keys, "&") || types.Int32(oldItem.Period) != period || oldItem.PeriodUnit != periodUnit || oldItem.Value != value {
		versionChanged = true
	}

	// 保存
	op := NewMetricItemOperator()
	op.Id = itemId
	op.Name = name
	if len(keys) > 0 {
		keysJSON, err := json.Marshal(keys)
		if err != nil {
			return err
		}
		op.Keys = keysJSON
	} else {
		op.Keys = "[]"
	}
	op.Period = period
	op.PeriodUnit = periodUnit
	op.Value = value
	op.IsOn = isOn
	if versionChanged {
		op.Version = dbs.SQL("version+1")
	}
	op.IsPublic = isPublic
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	// 通知更新
	if versionChanged || (oldItem.IsOn == 0 && isOn) || (oldItem.IsOn == 1 && !isOn) || oldIsPublic != isPublic {
		err := this.NotifyUpdate(tx, itemId, isPublic || oldIsPublic)
		if err != nil {
			return err
		}
	}

	// 删除旧数据
	if versionChanged {
		err := SharedMetricStatDAO.DeleteOldVersionItemStats(tx, itemId, types.Int32(oldItem.Version+1))
		if err != nil {
			return err
		}
	}

	return nil
}

// CountEnabledItems 计算指标的数量
func (this *MetricItemDAO) CountEnabledItems(tx *dbs.Tx, category serverconfigs.MetricItemCategory) (int64, error) {
	return this.Query(tx).
		State(MetricItemStateEnabled).
		Attr("userId", 0).
		Attr("category", category).
		Count()
}

// ListEnabledItems 列出单页指标
func (this *MetricItemDAO) ListEnabledItems(tx *dbs.Tx, category serverconfigs.MetricItemCategory, offset int64, size int64) (result []*MetricItem, err error) {
	_, err = this.Query(tx).
		State(MetricItemStateEnabled).
		Attr("userId", 0).
		Attr("category", category).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllPublicItems 取得公用的指标
func (this *MetricItemDAO) FindAllPublicItems(tx *dbs.Tx, category string, cacheMap *utils.CacheMap) (result []*MetricItem, err error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":FindAllPublicItems:" + category
	cache, ok := cacheMap.Get(cacheKey)
	if ok {
		return cache.([]*MetricItem), nil
	}

	_, err = this.Query(tx).
		State(MetricItemStateEnabled).
		Attr("userId", 0).
		Attr("category", category).
		Attr("isPublic", true).
		DescPk().
		Slice(&result).
		FindAll()
	if err != nil {
		return nil, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, result)
	}

	return
}

// ComposeItemConfig 组合指标配置
func (this *MetricItemDAO) ComposeItemConfig(tx *dbs.Tx, itemId int64) (*serverconfigs.MetricItemConfig, error) {
	if itemId <= 0 {
		return nil, nil
	}
	one, err := this.Query(tx).
		Pk(itemId).
		State(MetricItemStateEnabled).
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	var item = one.(*MetricItem)
	var config = &serverconfigs.MetricItemConfig{
		Id:         int64(item.Id),
		IsOn:       item.IsOn == 1,
		Period:     types.Int(item.Period),
		PeriodUnit: item.PeriodUnit,
		Category:   item.Category,
		Value:      item.Value,
		Keys:       item.DecodeKeys(),
		Version:    types.Int32(item.Version),
	}

	return config, nil
}

// ComposeItemConfigWithItem 根据Item信息组合指标
func (this *MetricItemDAO) ComposeItemConfigWithItem(item *MetricItem) *serverconfigs.MetricItemConfig {
	if item == nil {
		return nil
	}
	var config = &serverconfigs.MetricItemConfig{
		Id:         int64(item.Id),
		IsOn:       item.IsOn == 1,
		Period:     types.Int(item.Period),
		PeriodUnit: item.PeriodUnit,
		Category:   item.Category,
		Value:      item.Value,
		Keys:       item.DecodeKeys(),
		Version:    types.Int32(item.Version),
	}

	return config
}

// FindItemVersion 获取指标的版本号
func (this *MetricItemDAO) FindItemVersion(tx *dbs.Tx, itemId int64) (int32, error) {
	version, err := this.Query(tx).
		Pk(itemId).
		Result("version").
		FindIntCol(0)
	if err != nil {
		return 0, err
	}
	return types.Int32(version), nil
}

// NotifyUpdate 通知更新
func (this *MetricItemDAO) NotifyUpdate(tx *dbs.Tx, itemId int64, isPublic bool) error {
	if isPublic {
		clusterIds, err := SharedNodeClusterDAO.FindAllEnableClusterIds(tx)
		if err != nil {
			return err
		}
		for _, clusterId := range clusterIds {
			err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, NodeTaskTypeConfigChanged)
			if err != nil {
				return err
			}
		}
		return nil
	}
	clusterIds, err := SharedNodeClusterMetricItemDAO.FindAllClusterIdsWithItemId(tx, itemId)
	if err != nil {
		return err
	}
	for _, clusterId := range clusterIds {
		err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, NodeTaskTypeConfigChanged)
		if err != nil {
			return err
		}
	}
	return nil
}
