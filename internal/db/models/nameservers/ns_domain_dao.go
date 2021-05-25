package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSDomainStateEnabled  = 1 // 已启用
	NSDomainStateDisabled = 0 // 已禁用
)

type NSDomainDAO dbs.DAO

func NewNSDomainDAO() *NSDomainDAO {
	return dbs.NewDAO(&NSDomainDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSDomains",
			Model:  new(NSDomain),
			PkName: "id",
		},
	}).(*NSDomainDAO)
}

var SharedNSDomainDAO *NSDomainDAO

func init() {
	dbs.OnReady(func() {
		SharedNSDomainDAO = NewNSDomainDAO()
	})
}

// EnableNSDomain 启用条目
func (this *NSDomainDAO) EnableNSDomain(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSDomainStateEnabled).
		Update()
	return err
}

// DisableNSDomain 禁用条目
func (this *NSDomainDAO) DisableNSDomain(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSDomainStateDisabled).
		Update()
	return err
}

// FindEnabledNSDomain 查找启用中的条目
func (this *NSDomainDAO) FindEnabledNSDomain(tx *dbs.Tx, id uint32) (*NSDomain, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSDomainStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSDomain), err
}

// FindNSDomainName 根据主键查找名称
func (this *NSDomainDAO) FindNSDomainName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}
