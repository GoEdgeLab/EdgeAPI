package acme

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ACMEProviderAccountStateEnabled  = 1 // 已启用
	ACMEProviderAccountStateDisabled = 0 // 已禁用
)

type ACMEProviderAccountDAO dbs.DAO

func NewACMEProviderAccountDAO() *ACMEProviderAccountDAO {
	return dbs.NewDAO(&ACMEProviderAccountDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeACMEProviderAccounts",
			Model:  new(ACMEProviderAccount),
			PkName: "id",
		},
	}).(*ACMEProviderAccountDAO)
}

var SharedACMEProviderAccountDAO *ACMEProviderAccountDAO

func init() {
	dbs.OnReady(func() {
		SharedACMEProviderAccountDAO = NewACMEProviderAccountDAO()
	})
}

// EnableACMEProviderAccount 启用条目
func (this *ACMEProviderAccountDAO) EnableACMEProviderAccount(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ACMEProviderAccountStateEnabled).
		Update()
	return err
}

// DisableACMEProviderAccount 禁用条目
func (this *ACMEProviderAccountDAO) DisableACMEProviderAccount(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ACMEProviderAccountStateDisabled).
		Update()
	return err
}

// FindEnabledACMEProviderAccount 查找启用中的条目
func (this *ACMEProviderAccountDAO) FindEnabledACMEProviderAccount(tx *dbs.Tx, id int64) (*ACMEProviderAccount, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ACMEProviderAccountStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ACMEProviderAccount), err
}

// FindACMEProviderAccountName 根据主键查找名称
func (this *ACMEProviderAccountDAO) FindACMEProviderAccountName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateAccount 创建账号
func (this *ACMEProviderAccountDAO) CreateAccount(tx *dbs.Tx, name string, providerCode string, eabKid string, eabKey string) (int64, error) {
	var op = NewACMEProviderAccountOperator()
	op.Name = name
	op.ProviderCode = providerCode
	op.EabKid = eabKid
	op.EabKey = eabKey

	op.IsOn = true
	op.State = ACMEProviderAccountStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateAccount 修改账号
func (this *ACMEProviderAccountDAO) UpdateAccount(tx *dbs.Tx, accountId int64, name string, eabKid string, eabKey string) error {
	if accountId <= 0 {
		return errors.New("invalid accountId")
	}
	var op = NewACMEProviderAccountOperator()
	op.Id = accountId
	op.Name = name
	op.EabKid = eabKid
	op.EabKey = eabKey
	return this.Save(tx, op)
}

// CountAllEnabledAccounts 计算账号数量
func (this *ACMEProviderAccountDAO) CountAllEnabledAccounts(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		Count()
}

// ListEnabledAccounts 查找单页账号
func (this *ACMEProviderAccountDAO) ListEnabledAccounts(tx *dbs.Tx, offset int64, size int64) (result []*ACMEProviderAccount, err error) {
	_, err = this.Query(tx).
		State(ACMEProviderAccountStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledAccountsWithProviderCode 根据服务商代号查找账号
func (this *ACMEProviderAccountDAO) FindAllEnabledAccountsWithProviderCode(tx *dbs.Tx, providerCode string) (result []*ACMEProviderAccount, err error) {
	_, err = this.Query(tx).
		State(ACMEProviderAccountStateEnabled).
		Attr("providerCode", providerCode).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
