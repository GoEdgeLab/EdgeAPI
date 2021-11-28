package accounts

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		go func() {
			// 自动支付账单任务
			var ticker = time.NewTicker(12 * time.Hour)
			for range ticker.C {
				if SharedUserAccountDAO.Instance != nil {
					err := SharedUserAccountDAO.Instance.RunTx(func(tx *dbs.Tx) error {
						return SharedUserAccountDAO.PayBills(tx)
					})
					if err != nil {
						remotelogs.Error("USER_ACCOUNT_DAO", "pay bills task failed: "+err.Error())
					}
				}
			}
		}()
	})
}

type UserAccountDAO dbs.DAO

func NewUserAccountDAO() *UserAccountDAO {
	return dbs.NewDAO(&UserAccountDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserAccounts",
			Model:  new(UserAccount),
			PkName: "id",
		},
	}).(*UserAccountDAO)
}

var SharedUserAccountDAO *UserAccountDAO

func init() {
	dbs.OnReady(func() {
		SharedUserAccountDAO = NewUserAccountDAO()
	})
}

// FindUserAccountWithUserId 根据用户ID查找用户账户
func (this *UserAccountDAO) FindUserAccountWithUserId(tx *dbs.Tx, userId int64) (*UserAccount, error) {
	if userId <= 0 {
		return nil, errors.New("invalid userId '" + types.String(userId) + "'")
	}

	// 用户是否存在
	user, err := models.SharedUserDAO.FindEnabledUser(tx, userId, nil)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid userId '" + types.String(userId) + "'")
	}

	account, err := this.Query(tx).
		Attr("userId", userId).
		Find()
	if err != nil {
		return nil, err
	}
	if account != nil {
		return account.(*UserAccount), nil
	}

	var op = NewUserAccountOperator()
	op.UserId = userId
	_, err = this.SaveInt64(tx, op)
	if err != nil {
		return nil, err
	}
	return this.FindUserAccountWithUserId(tx, userId)
}

// FindUserAccountWithAccountId 根据ID查找用户账户
func (this *UserAccountDAO) FindUserAccountWithAccountId(tx *dbs.Tx, accountId int64) (*UserAccount, error) {
	one, err := this.Query(tx).
		Pk(accountId).
		Find()
	if one != nil {
		return one.(*UserAccount), nil
	}
	return nil, err
}

// UpdateUserAccount 操作用户账户
func (this *UserAccountDAO) UpdateUserAccount(tx *dbs.Tx, accountId int64, delta float32, eventType userconfigs.AccountEventType, description string, params maps.Map) error {
	account, err := this.FindUserAccountWithAccountId(tx, accountId)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("invalid account id '" + types.String(accountId) + "'")
	}
	var userId = int64(account.UserId)
	var deltaFloat64 = float64(delta)
	if deltaFloat64 < 0 && account.Total < -deltaFloat64 {
		return errors.New("not enough account quota to decrease")
	}

	// 操作账户
	err = this.Query(tx).
		Pk(account.Id).
		Set("total", dbs.SQL("total+:delta")).
		Param("delta", delta).
		UpdateQuickly()
	if err != nil {
		return err
	}

	// 生成日志
	err = SharedUserAccountLogDAO.CreateAccountLog(tx, userId, accountId, delta, 0, eventType, description, params)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserAccountFrozen 操作用户账户冻结余额
func (this *UserAccountDAO) UpdateUserAccountFrozen(tx *dbs.Tx, userId int64, delta float32, eventType userconfigs.AccountEventType, description string, params maps.Map) error {
	account, err := this.FindUserAccountWithUserId(tx, userId)
	if err != nil {
		return err
	}
	var deltaFloat64 = float64(delta)
	if deltaFloat64 < 0 && account.TotalFrozen < -deltaFloat64 {
		return errors.New("not enough account frozen quota to decrease")
	}

	// 操作账户
	err = this.Query(tx).
		Pk(account.Id).
		Set("totalFrozen", dbs.SQL("total+:delta")).
		Param("delta", delta).
		UpdateQuickly()
	if err != nil {
		return err
	}

	// 生成日志
	err = SharedUserAccountLogDAO.CreateAccountLog(tx, userId, int64(account.Id), 0, delta, eventType, description, params)
	if err != nil {
		return err
	}

	return nil
}

// CountAllAccounts 计算所有账户数量
func (this *UserAccountDAO) CountAllAccounts(tx *dbs.Tx, keyword string) (int64, error) {
	var query = this.Query(tx)
	if len(keyword) > 0 {
		query.Where("userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1 AND (username LIKE :keyword OR fullname LIKE :keyword))")
		query.Param("keyword", keyword)
	} else {
		query.Where("userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1)")
	}
	return query.Count()
}

// ListAccounts 列出单页账户
func (this *UserAccountDAO) ListAccounts(tx *dbs.Tx, keyword string, offset int64, size int64) (result []*UserAccount, err error) {
	var query = this.Query(tx)
	if len(keyword) > 0 {
		query.Where("userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1 AND (username LIKE :keyword OR fullname LIKE :keyword))")
		query.Param("keyword", keyword)
	} else {
		query.Where("userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1)")
	}
	_, err = query.
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// PayBills 尝试自动支付账单
func (this *UserAccountDAO) PayBills(tx *dbs.Tx) error {
	bills, err := models.SharedUserBillDAO.FindUnpaidBills(tx, 10000)
	if err != nil {
		return err
	}

	// 先支付久远的
	lists.Reverse(bills)

	for _, bill := range bills {
		if bill.Amount <= 0 {
			err = models.SharedUserBillDAO.UpdateUserBillIsPaid(tx, int64(bill.Id), true)
			if err != nil {
				return err
			}
			continue
		}

		account, err := SharedUserAccountDAO.FindUserAccountWithUserId(tx, int64(bill.UserId))
		if err != nil {
			return err
		}
		if account == nil || account.Total < bill.Amount {
			continue
		}

		// 扣款
		err = SharedUserAccountDAO.UpdateUserAccount(tx, int64(account.Id), -float32(bill.Amount), userconfigs.AccountEventTypePayBill, "支付账单"+bill.Code, maps.Map{"billId": bill.Id})
		if err != nil {
			return err
		}

		// 改为已支付
		err = models.SharedUserBillDAO.UpdateUserBillIsPaid(tx, int64(bill.Id), true)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckUserAccount 检查用户账户
func (this *UserAccountDAO) CheckUserAccount(tx *dbs.Tx, userId int64, accountId int64) error {
	exists, err := this.Query(tx).
		Pk(accountId).
		Attr("userId", userId).
		Exist()
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrNotFound
	}
	return nil
}
