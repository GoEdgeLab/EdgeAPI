//go:build !plus

package models

import (
	"encoding/json"
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
		AscPk().
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
