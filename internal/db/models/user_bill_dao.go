package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
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

		go func() {
			// 自动生成账单任务
			var ticker = time.NewTicker(1 * time.Minute)
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
		}()
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
func (this *UserBillDAO) CreateBill(tx *dbs.Tx, userId int64, billType BillType, description string, amount float32, month string) error {
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
			"isPaid":      false,
		}, maps.Map{
			"amount": amount,
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
	if len(regions) == 0 {
		return nil
	}

	priceItems, err := SharedNodePriceItemDAO.FindAllEnabledRegionPrices(tx, NodePriceTypeTraffic)
	if err != nil {
		return err
	}
	if len(priceItems) == 0 {
		return nil
	}

	// 计算套餐费用
	plans, err := SharedPlanDAO.FindAllEnabledPlans(tx)
	if err != nil {
		return err
	}
	var planMap = map[int64]*Plan{}
	for _, plan := range plans {
		planMap[int64(plan.Id)] = plan
	}

	stats, err := SharedServerDailyStatDAO.FindMonthlyStatsWithPlan(tx, month)
	if err != nil {
		return err
	}
	for _, stat := range stats {
		plan, ok := planMap[int64(stat.PlanId)]
		if !ok {
			continue
		}
		if plan.PriceType != serverconfigs.PlanPriceTypeTraffic {
			continue
		}
		if len(plan.TrafficPrice) == 0 {
			continue
		}
		var priceConfig = &serverconfigs.PlanTrafficPrice{}
		err = json.Unmarshal([]byte(plan.TrafficPrice), priceConfig)
		if err != nil {
			return err
		}
		if priceConfig.Base > 0 {
			var fee = priceConfig.Base * (float32(stat.Bytes) / 1024 / 1024 / 1024)
			err = SharedServerDailyStatDAO.UpdateStatFee(tx, int64(stat.Id), fee)
			if err != nil {
				return err
			}
		}
	}

	// 用户
	offset := int64(0)
	size := int64(100) // 每次只查询N次，防止由于执行时间过长而锁表
	for {
		userIds, err := SharedUserDAO.ListEnabledUserIds(tx, offset, size)
		if err != nil {
			return err
		}
		offset += size
		if len(userIds) == 0 {
			break
		}

		for _, userId := range userIds {
			// CDN流量账单
			err := this.generateTrafficBill(tx, userId, month, regions, priceItems)
			if err != nil {
				return err
			}
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

// 生成CDN流量账单
// month 格式YYYYMM
func (this *UserBillDAO) generateTrafficBill(tx *dbs.Tx, userId int64, month string, regions []*NodeRegion, priceItems []*NodePriceItem) error {
	// 检查是否已经有账单了
	if month < timeutil.Format("Ym") {
		b, err := this.ExistBill(tx, userId, BillTypeTraffic, month)
		if err != nil {
			return err
		}
		if b {
			return nil
		}
	}

	var cost = float32(0)
	for _, region := range regions {
		if len(region.Prices) == 0 || region.Prices == "null" {
			continue
		}
		priceMap := map[string]float32{}
		err := json.Unmarshal([]byte(region.Prices), &priceMap)
		if err != nil {
			return err
		}

		trafficBytes, err := SharedServerDailyStatDAO.SumUserMonthlyWithoutPlan(tx, userId, int64(region.Id), month)
		if err != nil {
			return err
		}
		if trafficBytes == 0 {
			continue
		}

		itemId := SharedNodePriceItemDAO.SearchItemsWithBytes(priceItems, trafficBytes)
		if itemId == 0 {
			continue
		}

		price, ok := priceMap[numberutils.FormatInt64(itemId)]
		if !ok {
			continue
		}

		// 计算钱
		// 这里采用1000进制
		cost += (float32(trafficBytes*8) / 1_000_000_000) * price
	}

	// 套餐费用
	planFee, err := SharedServerDailyStatDAO.SumUserMonthlyFee(tx, userId, month)
	if err != nil {
		return err
	}
	cost += float32(planFee)

	if cost == 0 {
		return nil
	}

	// 创建账单
	return this.CreateBill(tx, userId, BillTypeTraffic, "按流量计费", cost, month)
}

// BillTypeName 获取账单类型名称
func (this *UserBillDAO) BillTypeName(billType BillType) string {
	switch billType {
	case BillTypeTraffic:
		return "流量"
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
