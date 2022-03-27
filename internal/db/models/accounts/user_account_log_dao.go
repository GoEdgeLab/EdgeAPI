package accounts

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

type UserAccountLogDAO dbs.DAO

func NewUserAccountLogDAO() *UserAccountLogDAO {
	return dbs.NewDAO(&UserAccountLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserAccountLogs",
			Model:  new(UserAccountLog),
			PkName: "id",
		},
	}).(*UserAccountLogDAO)
}

var SharedUserAccountLogDAO *UserAccountLogDAO

func init() {
	dbs.OnReady(func() {
		SharedUserAccountLogDAO = NewUserAccountLogDAO()
	})
}

// CreateAccountLog 生成用户账户日志
func (this *UserAccountLogDAO) CreateAccountLog(tx *dbs.Tx, userId int64, accountId int64, delta float32, deltaFrozen float32, eventType userconfigs.AccountEventType, description string, params maps.Map) error {
	var op = NewUserAccountLogOperator()
	op.UserId = userId
	op.AccountId = accountId
	op.Delta = delta
	op.DeltaFrozen = deltaFrozen

	account, err := SharedUserAccountDAO.FindUserAccountWithAccountId(tx, accountId)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("invalid account id '" + types.String(accountId) + "'")
	}
	op.Total = account.Total
	op.TotalFrozen = account.TotalFrozen

	op.EventType = eventType
	op.Description = description

	if params == nil {
		params = maps.Map{}
	}
	op.Params = params.AsJSON()

	op.Day = timeutil.Format("Ymd")
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	return SharedUserAccountDailyStatDAO.UpdateDailyStat(tx)
}

// CountAccountLogs 计算日志数量
func (this *UserAccountLogDAO) CountAccountLogs(tx *dbs.Tx, userId int64, accountId int64, keyword string, eventType string) (int64, error) {
	var query = this.Query(tx)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if accountId > 0 {
		query.Attr("accountId", accountId)
	}
	if len(keyword) > 0 {
		query.Where("(userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1 AND (username LIKE :keyword OR fullname LIKE :keyword)) OR description LIKE :keyword)")
		query.Param("keyword", dbutils.QuoteLike(keyword))
	}
	if len(eventType) > 0 {
		query.Attr("eventType", eventType)
	}
	return query.Count()
}

// ListAccountLogs 列出单页日志
func (this *UserAccountLogDAO) ListAccountLogs(tx *dbs.Tx, userId int64, accountId int64, keyword string, eventType string, offset int64, size int64) (result []*UserAccountLog, err error) {
	var query = this.Query(tx)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if accountId > 0 {
		query.Attr("accountId", accountId)
	}
	if len(keyword) > 0 {
		query.Where("(userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1 AND (username LIKE :keyword OR fullname LIKE :keyword)) OR description LIKE :keyword)")
		query.Param("keyword", dbutils.QuoteLike(keyword))
	}
	if len(eventType) > 0 {
		query.Attr("eventType", eventType)
	}
	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// SumDailyEventTypes 统计某天数据总和
func (this *UserAccountLogDAO) SumDailyEventTypes(tx *dbs.Tx, day string, eventTypes []userconfigs.AccountEventType) (float32, error) {
	if len(eventTypes) == 0 {
		return 0, nil
	}
	result, err := this.Query(tx).
		Attr("day", day).
		Attr("eventType", eventTypes).
		Sum("delta", 0)
	if err != nil {
		return 0, err
	}
	return types.Float32(result), nil
}
