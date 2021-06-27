package metrics

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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
func (this *MetricItemDAO) DisableMetricItem(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MetricItemStateDisabled).
		Update()
	return err
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
func (this *MetricItemDAO) CreateItem(tx *dbs.Tx, code string, category string, name string, keys []string, period int32, periodUnit string, value string) (int64, error) {
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
	op.IsOn = true
	op.State = MetricItemStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateItem 修改\指标
func (this *MetricItemDAO) UpdateItem(tx *dbs.Tx, itemId int64, name string, keys []string, period int32, periodUnit string, value string, isOn bool) error {
	if itemId <= 0 {
		return errors.New("invalid itemId")
	}
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
	return this.Save(tx, op)
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
	}

	return config, nil
}
