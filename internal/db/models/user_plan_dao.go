//go:build !plus

package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	UserPlanStateEnabled  = 1 // 已启用
	UserPlanStateDisabled = 0 // 已禁用

	DefaultUserPlanMaxDay = "3000-01-01"
)

type UserPlanDAO dbs.DAO

func NewUserPlanDAO() *UserPlanDAO {
	return dbs.NewDAO(&UserPlanDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserPlans",
			Model:  new(UserPlan),
			PkName: "id",
		},
	}).(*UserPlanDAO)
}

var SharedUserPlanDAO *UserPlanDAO

func init() {
	dbs.OnReady(func() {
		SharedUserPlanDAO = NewUserPlanDAO()
	})
}

// EnableUserPlan 启用条目
func (this *UserPlanDAO) EnableUserPlan(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserPlanStateEnabled).
		Update()
	return err
}

// DisableUserPlan 禁用条目
func (this *UserPlanDAO) DisableUserPlan(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserPlanStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, id)
}

// FindEnabledUserPlan 查找启用中的条目
func (this *UserPlanDAO) FindEnabledUserPlan(tx *dbs.Tx, userPlanId int64, cacheMap *utils.CacheMap) (*UserPlan, error) {
	var cacheKey = this.Table + ":FindEnabledUserPlan:" + types.String(userPlanId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(*UserPlan), nil
		}
	}

	result, err := this.Query(tx).
		Pk(userPlanId).
		Attr("state", UserPlanStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, result)
	}

	return result.(*UserPlan), err
}

// FindUserPlanWithoutState 查找套餐，并不检查状态
// 防止因为删除套餐而导致计费失败
func (this *UserPlanDAO) FindUserPlanWithoutState(tx *dbs.Tx, userPlanId int64, cacheMap *utils.CacheMap) (*UserPlan, error) {
	var cacheKey = this.Table + ":FindUserPlanWithoutState:" + types.String(userPlanId)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(*UserPlan), nil
		}
	}

	result, err := this.Query(tx).
		Pk(userPlanId).
		Find()
	if result == nil {
		return nil, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, result)
	}

	return result.(*UserPlan), err
}

// CountAllEnabledUserPlans 计算套餐数量
func (this *UserPlanDAO) CountAllEnabledUserPlans(tx *dbs.Tx, userId int64, isAvailable bool, isExpired bool, expiringDays int32) (int64, error) {
	return 0, nil
}

// ListEnabledUserPlans 列出单页套餐
func (this *UserPlanDAO) ListEnabledUserPlans(tx *dbs.Tx, userId int64, isAvailable bool, isExpired bool, expiringDays int32, offset int64, size int64) (result []*UserPlan, err error) {
	return
}

// CreateUserPlan 创建套餐
func (this *UserPlanDAO) CreateUserPlan(tx *dbs.Tx, userId int64, planId int64, name string, dayTo string) (int64, error) {
	var op = NewUserPlanOperator()
	op.UserId = userId
	op.PlanId = planId
	op.Name = name
	op.DayTo = dayTo
	op.IsOn = true
	op.State = UserStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateUserPlan 修改套餐
func (this *UserPlanDAO) UpdateUserPlan(tx *dbs.Tx, userPlanId int64, planId int64, name string, dayTo string, isOn bool) error {
	if userPlanId <= 0 {
		return errors.New("invalid userPlanId")
	}
	var op = NewUserPlanOperator()
	op.Id = userPlanId
	op.PlanId = planId
	op.Name = name
	op.DayTo = dayTo
	op.IsOn = isOn
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, userPlanId)
}

// UpdateUserPlanDayTo 修改套餐日期
func (this *UserPlanDAO) UpdateUserPlanDayTo(tx *dbs.Tx, userPlanId int64, dayTo string) error {
	if userPlanId <= 0 {
		return errors.New("invalid userPlanId")
	}
	var op = NewUserPlanOperator()
	op.Id = userPlanId
	op.DayTo = dayTo
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, userPlanId)
}

// FindAllEnabledPlansForServer 列出服务可用的套餐
func (this *UserPlanDAO) FindAllEnabledPlansForServer(tx *dbs.Tx, userId int64, serverId int64) (result []*UserPlan, err error) {
	return
}

// CheckUserPlan 检查用户套餐
func (this *UserPlanDAO) CheckUserPlan(tx *dbs.Tx, userId int64, userPlanId int64) error {
	exists, err := this.Query(tx).
		Pk(userPlanId).
		Attr("userId", userId).
		State(UserPlanStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

// NotifyUpdate 通知更新
func (this *UserPlanDAO) NotifyUpdate(tx *dbs.Tx, userPlanId int64) error {
	serverId, err := SharedServerDAO.FindEnabledServerIdWithUserPlanId(tx, userPlanId)
	if err != nil {
		return err
	}
	if serverId > 0 {
		return SharedServerDAO.NotifyUpdate(tx, serverId)
	}
	return nil
}
