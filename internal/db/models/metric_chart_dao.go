package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

const (
	MetricChartStateEnabled  = 1 // 已启用
	MetricChartStateDisabled = 0 // 已禁用
)

type MetricChartDAO dbs.DAO

func NewMetricChartDAO() *MetricChartDAO {
	return dbs.NewDAO(&MetricChartDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMetricCharts",
			Model:  new(MetricChart),
			PkName: "id",
		},
	}).(*MetricChartDAO)
}

var SharedMetricChartDAO *MetricChartDAO

func init() {
	dbs.OnReady(func() {
		SharedMetricChartDAO = NewMetricChartDAO()
	})
}

// EnableMetricChart 启用条目
func (this *MetricChartDAO) EnableMetricChart(tx *dbs.Tx, chartId int64) error {
	_, err := this.Query(tx).
		Pk(chartId).
		Set("state", MetricChartStateEnabled).
		Update()
	return err
}

// DisableMetricChart 禁用条目
func (this *MetricChartDAO) DisableMetricChart(tx *dbs.Tx, chartId int64) error {
	_, err := this.Query(tx).
		Pk(chartId).
		Set("state", MetricChartStateDisabled).
		Update()
	return err
}

// FindEnabledMetricChart 查找启用中的条目
func (this *MetricChartDAO) FindEnabledMetricChart(tx *dbs.Tx, chartId int64) (*MetricChart, error) {
	result, err := this.Query(tx).
		Pk(chartId).
		Attr("state", MetricChartStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MetricChart), err
}

// FindMetricChartName 根据主键查找名称
func (this *MetricChartDAO) FindMetricChartName(tx *dbs.Tx, chartId int64) (string, error) {
	return this.Query(tx).
		Pk(chartId).
		Result("name").
		FindStringCol("")
}

// CreateChart 创建图表
func (this *MetricChartDAO) CreateChart(tx *dbs.Tx, itemId int64, name string, chartType string, widthDiv int32, maxItems int32, params maps.Map, ignoreEmptyKeys bool, ignoredKeys []string) (int64, error) {
	op := NewMetricChartOperator()
	op.ItemId = itemId
	op.Name = name
	op.Type = chartType
	op.WidthDiv = widthDiv
	op.MaxItems = maxItems

	if params == nil {
		params = maps.Map{}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return 0, err
	}
	op.Params = paramsJSON
	op.IgnoreEmptyKeys = ignoreEmptyKeys

	if len(ignoredKeys) == 0 {
		op.IgnoredKeys = "[]"
	} else {
		ignoredKeysJSON, err := json.Marshal(ignoredKeys)
		if err != nil {
			return 0, err
		}
		op.IgnoredKeys = ignoredKeysJSON
	}

	op.IsOn = true
	op.State = MetricChartStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateChart 修改图表
func (this *MetricChartDAO) UpdateChart(tx *dbs.Tx, chartId int64, name string, chartType string, widthDiv int32, maxItems int32, params maps.Map, ignoreEmptyKeys bool, ignoredKeys []string, isOn bool) error {
	if chartId <= 0 {
		return errors.New("invalid chartId")
	}
	op := NewMetricChartOperator()
	op.Id = chartId
	op.Name = name
	op.Type = chartType
	op.WidthDiv = widthDiv
	op.MaxItems = maxItems

	if params == nil {
		params = maps.Map{}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}
	op.Params = paramsJSON
	op.IgnoreEmptyKeys = ignoreEmptyKeys

	if len(ignoredKeys) == 0 {
		op.IgnoredKeys = "[]"
	} else {
		ignoredKeysJSON, err := json.Marshal(ignoredKeys)
		if err != nil {
			return err
		}
		op.IgnoredKeys = ignoredKeysJSON
	}

	op.IsOn = isOn

	return this.Save(tx, op)
}

// CountEnabledCharts 计算图表数量
func (this *MetricChartDAO) CountEnabledCharts(tx *dbs.Tx, itemId int64) (int64, error) {
	return this.Query(tx).
		Attr("itemId", itemId).
		State(MetricChartStateEnabled).
		Count()
}

// ListEnabledCharts 列出单页图表
func (this *MetricChartDAO) ListEnabledCharts(tx *dbs.Tx, itemId int64, offset int64, size int64) (result []*MetricChart, err error) {
	_, err = this.Query(tx).
		Attr("itemId", itemId).
		State(MetricChartStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledCharts 查找所有图表
func (this *MetricChartDAO) FindAllEnabledCharts(tx *dbs.Tx, itemId int64) (result []*MetricChart, err error) {
	_, err = this.Query(tx).
		Attr("itemId", itemId).
		State(MetricChartStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
