package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		var generatedMonths = []string{}

		goman.New(func() {
			// 自动生成账单任务
			var ticker = time.NewTicker(1 * time.Hour)
			for range ticker.C {
				// 是否已经生成了，如果已经生成了就跳过
				var lastMonth = timeutil.Format("Ym", time.Now().AddDate(0, -1, 0))
				if lists.ContainsString(generatedMonths, lastMonth) {
					continue
				}

				err := SharedUserBillDAO.GenerateBills(nil, lastMonth)
				if err != nil {
					remotelogs.Error("UserBillDAO", "generate bills failed: "+err.Error())
				} else {
					generatedMonths = append(generatedMonths, lastMonth)
				}
			}
		})
	})
}

type BillType = string

const (
	BillTypeTraffic BillType = "traffic" // 按流量计费
)

type UserBillDAO dbs.DAO

func NewUserBillDAO() *UserBillDAO {
	return dbs.NewDAO(&UserBillDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserBills",
			Model:  new(UserBill),
			PkName: "id",
		},
	}).(*UserBillDAO)
}

var SharedUserBillDAO *UserBillDAO

func init() {
	dbs.OnReady(func() {
		SharedUserBillDAO = NewUserBillDAO()
	})
}

// FindUserBill 查找单个账单
func (this *UserBillDAO) FindUserBill(tx *dbs.Tx, billId int64) (*UserBill, error) {
	one, err := this.Query(tx).
		Pk(billId).
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*UserBill), nil
}

// CountAllUserBills 计算账单数量
func (this *UserBillDAO) CountAllUserBills(tx *dbs.Tx, isPaid int32, userId int64, month string) (int64, error) {
	query := this.Query(tx)
	if isPaid == 0 {
		query.Attr("isPaid", 0)
	} else if isPaid > 0 {
		query.Attr("isPaid", 1)
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if len(month) > 0 {
		query.Attr("month", month)
	}
	return query.Count()
}

// ListUserBills 列出单页账单
func (this *UserBillDAO) ListUserBills(tx *dbs.Tx, isPaid int32, userId int64, month string, offset, size int64) (result []*UserBill, err error) {
	query := this.Query(tx)
	if isPaid == 0 {
		query.Attr("isPaid", 0)
	} else if isPaid > 0 {
		query.Attr("isPaid", 1)
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if len(month) > 0 {
		query.Attr("month", month)
	}
	_, err = query.
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// FindUnpaidBills 查找未支付订单
func (this *UserBillDAO) FindUnpaidBills(tx *dbs.Tx, size int64) (result []*UserBill, err error) {
	if size <= 0 {
		size = 10000
	}
	_, err = this.Query(tx).
		Attr("isPaid", false).
		Lt("month", timeutil.Format("Ym")). //当月的不能支付，因为当月还没过完
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CreateBill 创建账单
func (this *UserBillDAO) CreateBill(tx *dbs.Tx, userId int64, billType BillType, description string, amount float64, month string, canPay bool) error {
	code, err := this.GenerateBillCode(tx)
	if err != nil {
		return err
	}
	return this.Query(tx).
		InsertOrUpdateQuickly(maps.Map{
			"userId":      userId,
			"type":        billType,
			"description": description,
			"amount":      amount,
			"month":       month,
			"code":        code,
			"isPaid":      amount == 0,
			"canPay":      canPay,
		}, maps.Map{
			"amount": amount,
			"canPay": canPay,
		})
}

// ExistBill 检查是否有当月账单
func (this *UserBillDAO) ExistBill(tx *dbs.Tx, userId int64, billType BillType, month string) (bool, error) {
	return this.Query(tx).
		Attr("userId", userId).
		Attr("month", month).
		Attr("type", billType).
		Exist()
}

// GenerateBills 生成账单
// month 格式YYYYMM
func (this *UserBillDAO) GenerateBills(tx *dbs.Tx, month string) error {
	// 区域价格
	regions, err := SharedNodeRegionDAO.FindAllEnabledRegionPrices(tx)
	if err != nil {
		return err
	}

	var priceItems []*NodePriceItem
	if len(regions) > 0 {
		priceItems, err = SharedNodePriceItemDAO.FindAllEnabledRegionPrices(tx, NodePriceTypeTraffic)
		if err != nil {
			return err
		}
	}

	// 默认计费方式
	userFinanceConfig, err := SharedSysSettingDAO.ReadUserFinanceConfig(tx)
	if err != nil {
		return err
	}

	// 计算服务套餐费用
	plans, err := SharedPlanDAO.FindAllEnabledPlans(tx)
	if err != nil {
		return err
	}
	var planMap = map[int64]*Plan{}
	for _, plan := range plans {
		planMap[int64(plan.Id)] = plan
	}

	var dayFrom = month + "01"
	var dayTo = month + "32"
	serverIds, err := SharedServerDailyStatDAO.FindDistinctServerIds(tx, dayFrom, dayTo)
	if err != nil {
		return err
	}
	var cacheMap = utils.NewCacheMap()
	var userIds = []int64{}
	for _, serverId := range serverIds {
		// 套餐类型
		userPlanId, userId, err := SharedServerDAO.FindServerLastUserPlanIdAndUserId(tx, serverId)
		if err != nil {
			return err
		}
		if userId == 0 {
			continue
		}

		userIds = append(userIds, userId)
		if userPlanId == 0 {
			// 总流量
			totalTrafficBytes, err := SharedServerDailyStatDAO.SumMonthlyBytes(tx, serverId, month)
			if err != nil {
				return err
			}

			// 默认计费方式
			if userFinanceConfig != nil && userFinanceConfig.IsOn { // 默认计费方式
				switch userFinanceConfig.PriceType {
				case serverconfigs.PlanPriceTypeTraffic:
					var config = userFinanceConfig.TrafficPriceConfig
					var fee float64 = 0
					if config != nil && config.Base > 0 {
						fee = float64(totalTrafficBytes) / 1024 / 1024 / 1024 * float64(config.Base)
					}

					// 百分位
					var percentile = 95
					percentileBytes, err := SharedServerDailyStatDAO.FindMonthlyPercentile(tx, serverId, month, percentile)
					if err != nil {
						return err
					}

					err = SharedServerBillDAO.CreateOrUpdateServerBill(tx, userId, serverId, month, userPlanId, 0, totalTrafficBytes, percentileBytes, percentile, userFinanceConfig.PriceType, fee)
					if err != nil {
						return err
					}
				case serverconfigs.PlanPriceTypeBandwidth:
					// 百分位
					var percentile = 95
					var config = userFinanceConfig.BandwidthPriceConfig
					if config != nil {
						percentile = config.Percentile
						if percentile <= 0 {
							percentile = 95
						} else if percentile > 100 {
							percentile = 100
						}
					}
					percentileBytes, err := SharedServerDailyStatDAO.FindMonthlyPercentile(tx, serverId, month, percentile)
					if err != nil {
						return err
					}
					var mb = float32(percentileBytes) / 1024 / 1024
					var price float32
					if config != nil {
						price = config.LookupPrice(mb)
					}
					var fee = float64(price)
					err = SharedServerBillDAO.CreateOrUpdateServerBill(tx, userId, serverId, month, userPlanId, 0, totalTrafficBytes, percentileBytes, percentile, userFinanceConfig.PriceType, fee)
					if err != nil {
						return err
					}
				}
			} else { // 区域流量计费
				var fee float64

				for _, region := range regions {
					var regionId = int64(region.Id)
					var pricesMap = region.DecodePriceMap()
					if len(pricesMap) == 0 {
						continue
					}

					trafficBytes, err := SharedServerDailyStatDAO.SumServerMonthlyWithRegion(tx, serverId, regionId, month)
					if err != nil {
						return err
					}
					if trafficBytes == 0 {
						continue
					}
					var itemId = SharedNodePriceItemDAO.SearchItemsWithBytes(priceItems, trafficBytes)
					if itemId == 0 {
						continue
					}
					price, ok := pricesMap[itemId]
					if !ok {
						continue
					}
					if price <= 0 {
						continue
					}
					var regionFee = float64(trafficBytes) / 1000 / 1000 / 1000 * 8 * price
					fee += regionFee
				}

				// 百分位
				var percentile = 95
				percentileBytes, err := SharedServerDailyStatDAO.FindMonthlyPercentile(tx, serverId, month, percentile)
				if err != nil {
					return err
				}

				err = SharedServerBillDAO.CreateOrUpdateServerBill(tx, userId, serverId, month, userPlanId, 0, totalTrafficBytes, percentileBytes, percentile, "", fee)
				if err != nil {
					return err
				}
			}
		} else {
			userPlan, err := SharedUserPlanDAO.FindUserPlanWithoutState(tx, userPlanId, cacheMap)
			if err != nil {
				return err
			}
			if userPlan == nil {
				continue
			}

			plan, ok := planMap[int64(userPlan.PlanId)]
			if !ok {
				continue
			}

			// 总流量
			totalTrafficBytes, err := SharedServerDailyStatDAO.SumMonthlyBytes(tx, serverId, month)
			if err != nil {
				return err
			}

			switch plan.PriceType {
			case serverconfigs.PlanPriceTypePeriod:
				// 已经在购买套餐的时候付过费，这里不再重复计费
				var fee float64 = 0

				// 百分位
				var percentile = 95
				percentileBytes, err := SharedServerDailyStatDAO.FindMonthlyPercentile(tx, serverId, month, percentile)
				if err != nil {
					return err
				}

				err = SharedServerBillDAO.CreateOrUpdateServerBill(tx, int64(userPlan.UserId), serverId, month, userPlanId, int64(userPlan.PlanId), totalTrafficBytes, percentileBytes, percentile, plan.PriceType, fee)
				if err != nil {
					return err
				}
			case serverconfigs.PlanPriceTypeTraffic:
				var config = plan.DecodeTrafficPrice()
				var fee float64 = 0
				if config != nil && config.Base > 0 {
					fee = float64(totalTrafficBytes) / 1024 / 1024 / 1024 * float64(config.Base)
				}

				// 百分位
				var percentile = 95
				percentileBytes, err := SharedServerDailyStatDAO.FindMonthlyPercentile(tx, serverId, month, percentile)
				if err != nil {
					return err
				}

				err = SharedServerBillDAO.CreateOrUpdateServerBill(tx, int64(userPlan.UserId), serverId, month, userPlanId, int64(userPlan.PlanId), totalTrafficBytes, percentileBytes, percentile, plan.PriceType, fee)
				if err != nil {
					return err
				}
			case serverconfigs.PlanPriceTypeBandwidth:
				// 百分位
				var percentile = 95
				var config = plan.DecodeBandwidthPrice()
				if config != nil {
					percentile = config.Percentile
					if percentile <= 0 {
						percentile = 95
					} else if percentile > 100 {
						percentile = 100
					}
				}
				percentileBytes, err := SharedServerDailyStatDAO.FindMonthlyPercentile(tx, serverId, month, percentile)
				if err != nil {
					return err
				}
				var mb = float32(percentileBytes) / 1024 / 1024
				var price float32
				if config != nil {
					price = config.LookupPrice(mb)
				}
				var fee = float64(price)
				err = SharedServerBillDAO.CreateOrUpdateServerBill(tx, int64(userPlan.UserId), serverId, month, userPlanId, int64(userPlan.PlanId), totalTrafficBytes, percentileBytes, percentile, plan.PriceType, fee)
				if err != nil {
					return err
				}
			}
		}
	}

	// 计算用户费用
	for _, userId := range userIds {
		if userId == 0 {
			continue
		}
		amount, err := SharedServerBillDAO.SumUserMonthlyAmount(tx, userId, month)
		if err != nil {
			return err
		}
		err = SharedUserBillDAO.CreateBill(tx, userId, BillTypeTraffic, "流量带宽费用", amount, month, month < timeutil.Format("Ym"))
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateUserBillIsPaid 设置账单已支付
func (this *UserBillDAO) UpdateUserBillIsPaid(tx *dbs.Tx, billId int64, isPaid bool) error {
	return this.Query(tx).
		Pk(billId).
		Set("isPaid", isPaid).
		UpdateQuickly()
}

// BillTypeName 获取账单类型名称
func (this *UserBillDAO) BillTypeName(billType BillType) string {
	switch billType {
	case BillTypeTraffic:
		return "流量带宽"
	}
	return ""
}

// GenerateBillCode 生成账单编号
func (this *UserBillDAO) GenerateBillCode(tx *dbs.Tx) (string, error) {
	var code = timeutil.Format("YmdHis") + types.String(rands.Int(100000, 999999))
	exists, err := this.Query(tx).
		Attr("code", code).
		Exist()
	if err != nil {
		return "", err
	}
	if !exists {
		return code, nil
	}
	return this.GenerateBillCode(tx)
}

// CheckUserBill 检查用户账单
func (this *UserBillDAO) CheckUserBill(tx *dbs.Tx, userId int64, billId int64) error {
	if userId <= 0 || billId <= 0 {
		return ErrNotFound
	}
	exists, err := this.Query(tx).
		Pk(billId).
		Attr("userId", userId).
		Exist()
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

// SumUnpaidUserBill 计算某个用户未支付总额
func (this *UserBillDAO) SumUnpaidUserBill(tx *dbs.Tx, userId int64) (float32, error) {
	sum, err := this.Query(tx).
		Attr("userId", userId).
		Attr("isPaid", 0).
		Sum("amount", 0)
	if err != nil {
		return 0, err
	}
	return float32(sum), nil
}
