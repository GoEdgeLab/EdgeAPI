package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"math"
	"time"
)

type ServerBillDAO dbs.DAO

func NewServerBillDAO() *ServerBillDAO {
	return dbs.NewDAO(&ServerBillDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerBills",
			Model:  new(ServerBill),
			PkName: "id",
		},
	}).(*ServerBillDAO)
}

var SharedServerBillDAO *ServerBillDAO

func init() {
	dbs.OnReady(func() {
		SharedServerBillDAO = NewServerBillDAO()
	})
}

// CreateOrUpdateServerBill 创建账单
func (this *ServerBillDAO) CreateOrUpdateServerBill(tx *dbs.Tx, userId int64, serverId int64, month string, userPlanId int64, planId int64, totalTrafficBytes int64, bandwidthPercentileBytes int64, bandwidthPercentile int, fee float64) error {
	fee = math.Floor(fee*100) / 100
	return this.Query(tx).
		InsertOrUpdateQuickly(maps.Map{
			"userId":                   userId,
			"serverId":                 serverId,
			"month":                    month,
			"amount":                   fee,
			"userPlanId":               userPlanId,
			"planId":                   planId,
			"totalTrafficBytes":        totalTrafficBytes,
			"bandwidthPercentileBytes": bandwidthPercentileBytes,
			"bandwidthPercentile":      bandwidthPercentile,
			"createdAt":                time.Now().Unix(),
		}, maps.Map{
			"userId":                   userId,
			"amount":                   fee,
			"userPlanId":               userPlanId,
			"planId":                   planId,
			"totalTrafficBytes":        totalTrafficBytes,
			"bandwidthPercentileBytes": bandwidthPercentileBytes,
			"bandwidthPercentile":      bandwidthPercentile,
			"createdAt":                time.Now().Unix(),
		})
}

// SumUserMonthlyAmount 计算总费用
func (this *ServerBillDAO) SumUserMonthlyAmount(tx *dbs.Tx, userId int64, month string) (float64, error) {
	return this.Query(tx).
		Attr("userId", userId).
		Attr("month", month).
		Sum("amount", 0)
}

// CountServerBills 计算总账单数量
func (this *ServerBillDAO) CountServerBills(tx *dbs.Tx, userId int64, month string) (int64, error) {
	var query = this.Query(tx)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if len(month) > 0 {
		query.Attr("month", month)
	}
	return query.Count()
}

// ListServerBills 列出单页账单
func (this *ServerBillDAO) ListServerBills(tx *dbs.Tx, userId int64, month string, offset int64, size int64) (result []*ServerBill, err error) {
	var query = this.Query(tx)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if len(month) > 0 {
		query.Attr("month", month)
	}
	_, err = query.
		Desc("serverId").
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}
