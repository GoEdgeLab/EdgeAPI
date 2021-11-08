package accounts

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

type UserAccountDailyStatDAO dbs.DAO

func NewUserAccountDailyStatDAO() *UserAccountDailyStatDAO {
	return dbs.NewDAO(&UserAccountDailyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserAccountDailyStats",
			Model:  new(UserAccountDailyStat),
			PkName: "id",
		},
	}).(*UserAccountDailyStatDAO)
}

var SharedUserAccountDailyStatDAO *UserAccountDailyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedUserAccountDailyStatDAO = NewUserAccountDailyStatDAO()
	})
}

// UpdateDailyStat 更新当天统计数据
func (this *UserAccountDailyStatDAO) UpdateDailyStat(tx *dbs.Tx) error {
	var day = timeutil.Format("Ymd")
	var month = timeutil.Format("Ym")
	income, err := SharedUserAccountLogDAO.SumDailyEventTypes(tx, day, userconfigs.AccountIncomeEventTypes)
	if err != nil {
		return err
	}

	expense, err := SharedUserAccountLogDAO.SumDailyEventTypes(tx, day, userconfigs.AccountExpenseEventTypes)
	if err != nil {
		return err
	}
	if expense < 0 {
		expense = -expense
	}

	return this.Query(tx).
		InsertOrUpdateQuickly(maps.Map{
			"day":     day,
			"month":   month,
			"income":  income,
			"expense": expense,
		}, maps.Map{
			"income":  income,
			"expense": expense,
		})
}

// FindDailyStats 查看按天统计
func (this *UserAccountDailyStatDAO) FindDailyStats(tx *dbs.Tx, dayFrom string, dayTo string) (result []*UserAccountDailyStat, err error) {
	_, err = this.Query(tx).
		Between("day", dayFrom, dayTo).
		Slice(&result).
		FindAll()
	return
}

// FindMonthlyStats 查看某月统计
func (this *UserAccountDailyStatDAO) FindMonthlyStats(tx *dbs.Tx, dayFrom string, dayTo string) (result []*UserAccountDailyStat, err error) {
	_, err = this.Query(tx).
		Result("SUM(income) AS income", "SUM(expense) AS expense", "month").
		Between("day", dayFrom, dayTo).
		Group("month").
		Slice(&result).
		FindAll()
	return
}
