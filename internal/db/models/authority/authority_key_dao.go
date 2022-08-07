package authority

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type AuthorityKeyDAO dbs.DAO

func NewAuthorityKeyDAO() *AuthorityKeyDAO {
	return dbs.NewDAO(&AuthorityKeyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeAuthorityKeys",
			Model:  new(AuthorityKey),
			PkName: "id",
		},
	}).(*AuthorityKeyDAO)
}

var SharedAuthorityKeyDAO *AuthorityKeyDAO

func init() {
	dbs.OnReady(func() {
		SharedAuthorityKeyDAO = NewAuthorityKeyDAO()

		// 初始化IsPlus值
		_, _ = SharedAuthorityKeyDAO.IsPlus(nil)
	})
}
