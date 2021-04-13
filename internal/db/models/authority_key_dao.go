package models

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"time"
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
	})
}

// UpdateKey 设置Key
func (this *AuthorityKeyDAO) UpdateKey(tx *dbs.Tx, value string, dayFrom string, dayTo string, hostname string, macAddresses []string) error {
	one, err := this.Query(tx).
		AscPk().
		Find()
	if err != nil {
		return err
	}
	op := NewAuthorityKeyOperator()
	if one != nil {
		op.Id = one.(*AuthorityKey).Id
	}
	op.Value = value
	op.DayFrom = dayFrom
	op.DayTo = dayTo
	op.Hostname = hostname

	if len(macAddresses) == 0 {
		macAddresses = []string{}
	}
	macAddressesJSON, err := json.Marshal(macAddresses)
	if err != nil {
		return err
	}

	op.MacAddresses = macAddressesJSON
	op.UpdatedAt = time.Now().Unix()

	return this.Save(tx, op)
}

// ReadKey 读取Key
func (this *AuthorityKeyDAO) ReadKey(tx *dbs.Tx) (key *AuthorityKey, err error) {
	one, err := this.Query(tx).
		AscPk().
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*AuthorityKey), nil
}

// ResetKey 重置Key
func (this *AuthorityKeyDAO) ResetKey(tx *dbs.Tx) error {
	_, err := this.Query(tx).
		Delete()
	return err
}
