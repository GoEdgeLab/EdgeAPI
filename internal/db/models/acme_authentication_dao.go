package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type ACMEAuthenticationDAO dbs.DAO

func NewACMEAuthenticationDAO() *ACMEAuthenticationDAO {
	return dbs.NewDAO(&ACMEAuthenticationDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeACMEAuthentications",
			Model:  new(ACMEAuthentication),
			PkName: "id",
		},
	}).(*ACMEAuthenticationDAO)
}

var SharedACMEAuthenticationDAO *ACMEAuthenticationDAO

func init() {
	dbs.OnReady(func() {
		SharedACMEAuthenticationDAO = NewACMEAuthenticationDAO()
	})
}

// 创建认证信息
func (this *ACMEAuthenticationDAO) CreateAuth(tx *dbs.Tx, taskId int64, domain string, token string, key string) error {
	op := NewACMEAuthenticationOperator()
	op.TaskId = taskId
	op.Domain = domain
	op.Token = token
	op.Key = key
	err := this.Save(tx, op)
	return err
}

// 根据令牌查找认证信息
func (this *ACMEAuthenticationDAO) FindAuthWithToken(tx *dbs.Tx, token string) (*ACMEAuthentication, error) {
	one, err := this.Query(tx).
		Attr("token", token).
		DescPk().
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*ACMEAuthentication), nil
}
