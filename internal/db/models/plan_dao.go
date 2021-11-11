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
	PlanStateEnabled  = 1 // 已启用
	PlanStateDisabled = 0 // 已禁用
)

type PlanDAO dbs.DAO

func NewPlanDAO() *PlanDAO {
	return dbs.NewDAO(&PlanDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgePlans",
			Model:  new(Plan),
			PkName: "id",
		},
	}).(*PlanDAO)
}

var SharedPlanDAO *PlanDAO

func init() {
	dbs.OnReady(func() {
		SharedPlanDAO = NewPlanDAO()
	})
}

// EnablePlan 启用条目
func (this *PlanDAO) EnablePlan(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", PlanStateEnabled).
		Update()
	return err
}

// DisablePlan 禁用条目
func (this *PlanDAO) DisablePlan(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", PlanStateDisabled).
		Update()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, id)
}

// FindEnabledPlan 查找启用中的条目
func (this *PlanDAO) FindEnabledPlan(tx *dbs.Tx, id int64) (*Plan, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", PlanStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Plan), err
}

// FindPlanName 根据主键查找名称
func (this *PlanDAO) FindPlanName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreatePlan 创建套餐
func (this *PlanDAO) CreatePlan(tx *dbs.Tx, name string, clusterId int64, trafficLimitJSON []byte, featuresJSON []byte, priceType serverconfigs.PlanPriceType, trafficPriceJSON []byte, monthlyPrice float32, seasonallyPrice float32, yearlyPrice float32) (int64, error) {
	var op = NewPlanOperator()
	op.Name = name
	op.ClusterId = clusterId
	if len(trafficLimitJSON) > 0 {
		op.TrafficLimit = trafficLimitJSON
	}
	if len(featuresJSON) > 0 {
		op.Features = featuresJSON
	}
	op.PriceType = priceType
	if len(trafficPriceJSON) > 0 {
		op.TrafficPrice = trafficPriceJSON
	}
	if monthlyPrice >= 0 {
		op.MonthlyPrice = monthlyPrice
	}
	if seasonallyPrice >= 0 {
		op.SeasonallyPrice = seasonallyPrice
	}
	if yearlyPrice >= 0 {
		op.YearlyPrice = yearlyPrice
	}
	op.IsOn = true
	op.State = PlanStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdatePlan 修改套餐
func (this *PlanDAO) UpdatePlan(tx *dbs.Tx, planId int64, name string, isOn bool, clusterId int64, trafficLimitJSON []byte, featuresJSON []byte, priceType serverconfigs.PlanPriceType, trafficPriceJSON []byte, monthlyPrice float32, seasonallyPrice float32, yearlyPrice float32) error {
	if planId <= 0 {
		return errors.New("invalid planId")
	}

	// 检查集群有无变化
	oldClusterId, err := this.Query(tx).
		Pk(planId).
		Result("clusterId").
		FindInt64Col(0)
	if err != nil {
		return err
	}

	var op = NewPlanOperator()
	op.Id = planId
	op.Name = name
	op.IsOn = isOn
	op.ClusterId = clusterId
	if len(trafficLimitJSON) > 0 {
		op.TrafficLimit = trafficLimitJSON
	}
	if len(featuresJSON) > 0 {
		op.Features = featuresJSON
	}
	op.PriceType = priceType
	if len(trafficPriceJSON) > 0 {
		op.TrafficPrice = trafficPriceJSON
	}
	if monthlyPrice >= 0 {
		op.MonthlyPrice = monthlyPrice
	} else {
		op.MonthlyPrice = 0
	}
	if seasonallyPrice >= 0 {
		op.SeasonallyPrice = seasonallyPrice
	} else {
		op.SeasonallyPrice = 0
	}
	if yearlyPrice >= 0 {
		op.YearlyPrice = yearlyPrice
	} else {
		op.YearlyPrice = 0
	}
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	if oldClusterId != clusterId {
		// 修改服务所属集群
		err = SharedServerDAO.UpdateServersClusterIdWithPlanId(tx, planId, clusterId)
		if err != nil {
			return err
		}

		err = SharedNodeClusterDAO.NotifyUpdate(tx, oldClusterId)
		if err != nil {
			return err
		}
	}

	return this.NotifyUpdate(tx, planId)
}

// CountAllEnabledPlans 计算套餐的数量
func (this *PlanDAO) CountAllEnabledPlans(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(PlanStateEnabled).
		Count()
}

// ListEnabledPlans 列出单页套餐
func (this *PlanDAO) ListEnabledPlans(tx *dbs.Tx, offset int64, size int64) (result []*Plan, err error) {
	_, err = this.Query(tx).
		State(PlanStateEnabled).
		Offset(offset).
		Limit(size).
		Slice(&result).
		Desc("order").
		DescPk().
		FindAll()
	return
}

// FindAllEnabledPlans 查找所有可用套餐
func (this *PlanDAO) FindAllEnabledPlans(tx *dbs.Tx) (result []*Plan, err error) {
	_, err = this.Query(tx).
		State(PlanStateEnabled).
		Slice(&result).
		FindAll()
	return
}

// SortPlans 增加排序
func (this *PlanDAO) SortPlans(tx *dbs.Tx, planIds []int64) error {
	if len(planIds) == 0 {
		return nil
	}
	var order = len(planIds)
	for _, planId := range planIds {
		err := this.Query(tx).
			Pk(planId).
			Set("order", order).
			UpdateQuickly()
		if err != nil {
			return err
		}
		order--
	}
	return nil
}

// FindEnabledPlanTrafficLimit 获取套餐的流量限制
func (this *PlanDAO) FindEnabledPlanTrafficLimit(tx *dbs.Tx, planId int64, cacheMap *utils.CacheMap) (*serverconfigs.TrafficLimitConfig, error) {
	var cacheKey = this.Table + ":FindEnabledPlanTrafficLimit:" + types.String(planId)
	if cacheMap != nil {
		cache, _ := cacheMap.Get(cacheKey)
		if cache != nil {
			return cache.(*serverconfigs.TrafficLimitConfig), nil
		}
	}

	trafficLimit, err := this.Query(tx).
		Pk(planId).
		State(PlanStateEnabled).
		Result("trafficLimit").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(trafficLimit) == 0 {
		return nil, nil
	}
	var config = &serverconfigs.TrafficLimitConfig{}
	err = json.Unmarshal([]byte(trafficLimit), config)
	if err != nil {
		return nil, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// NotifyUpdate 通知变更
func (this *PlanDAO) NotifyUpdate(tx *dbs.Tx, planId int64) error {
	// 这里不要加入状态参数，因为需要适应删除后的更新
	clusterId, err := this.Query(tx).
		Pk(planId).
		Result("clusterId").
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if clusterId > 0 {
		return SharedNodeClusterDAO.NotifyUpdate(tx, clusterId)
	}
	return nil
}
