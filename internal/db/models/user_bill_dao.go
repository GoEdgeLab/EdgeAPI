package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

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

// 计算账单数量
func (this *UserBillDAO) CountAllUserBills(isPaid int32, userId int64, month string) (int64, error) {
	query := this.Query()
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

// 列出单页账单
func (this *UserBillDAO) ListUserBills(isPaid int32, userId int64, month string, offset, size int64) (result []*UserBill, err error) {
	query := this.Query()
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

// 创建账单
func (this *UserBillDAO) CreateBill(userId int64, billType BillType, description string, amount float32, month string) (int64, error) {
	op := NewUserBillOperator()
	op.UserId = userId
	op.Type = billType
	op.Description = description
	op.Amount = amount
	op.Month = month
	op.IsPaid = false
	return this.SaveInt64(op)
}

// 检查是否有当月账单
func (this *UserBillDAO) ExistBill(userId int64, billType BillType, month string) (bool, error) {
	return this.Query().
		Attr("userId", userId).
		Attr("month", month).
		Attr("type", billType).
		Exist()
}

// 生成账单
// month 格式YYYYMM
func (this *UserBillDAO) GenerateBills(month string) error {
	// 用户
	offset := int64(0)
	size := int64(100) // 每次只查询N次，防止由于执行时间过长而锁表
	for {
		userIds, err := SharedUserDAO.ListEnabledUserIds(offset, size)
		if err != nil {
			return err
		}
		offset += size
		if len(userIds) == 0 {
			break
		}

		for _, userId := range userIds {
			// CDN流量账单
			err := this.generateTrafficBill(userId, month)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 生成CDN流量账单
// month 格式YYYYMM
func (this *UserBillDAO) generateTrafficBill(userId int64, month string) error {
	// 检查是否已经有账单了
	b, err := this.ExistBill(userId, BillTypeTraffic, month)
	if err != nil {
		return err
	}
	if b {
		return nil
	}

	// TODO 优化使用缓存
	regions, err := SharedNodeRegionDAO.FindAllEnabledRegionPrices()
	if err != nil {
		return err
	}
	if len(regions) == 0 {
		return nil
	}

	priceItems, err := SharedNodePriceItemDAO.FindAllEnabledRegionPrices(NodePriceTypeTraffic)
	if err != nil {
		return err
	}
	if len(priceItems) == 0 {
		return nil
	}

	cost := float32(0)
	for _, region := range regions {
		if len(region.Prices) == 0 || region.Prices == "null" {
			continue
		}
		priceMap := map[string]float32{}
		err = json.Unmarshal([]byte(region.Prices), &priceMap)
		if err != nil {
			return err
		}

		trafficBytes, err := SharedServerDailyStatDAO.SumUserMonthly(userId, int64(region.Id), month)
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

	if cost == 0 {
		return nil
	}

	// 创建账单
	_, err = this.CreateBill(userId, BillTypeTraffic, "按流量计费", cost, month)
	return err
}

// 获取账单类型名称
func (this *UserBillDAO) BillTypeName(billType BillType) string {
	switch billType {
	case BillTypeTraffic:
		return "流量"
	}
	return ""
}
