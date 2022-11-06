//go:build !plus

package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSRecordStateEnabled  = 1 // 已启用
	NSRecordStateDisabled = 0 // 已禁用
)

type NSRecordDAO dbs.DAO

func NewNSRecordDAO() *NSRecordDAO {
	return dbs.NewDAO(&NSRecordDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSRecords",
			Model:  new(NSRecord),
			PkName: "id",
		},
	}).(*NSRecordDAO)
}

var SharedNSRecordDAO *NSRecordDAO

func init() {
	dbs.OnReady(func() {
		SharedNSRecordDAO = NewNSRecordDAO()
	})
}

// EnableNSRecord 启用条目
func (this *NSRecordDAO) EnableNSRecord(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSRecordStateEnabled).
		Update()
	return err
}

// DisableNSRecord 禁用条目
func (this *NSRecordDAO) DisableNSRecord(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSRecordStateDisabled).
		Update()
	return err
}

// FindEnabledNSRecord 查找启用中的条目
func (this *NSRecordDAO) FindEnabledNSRecord(tx *dbs.Tx, id uint64) (*NSRecord, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(NSRecordStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSRecord), err
}

// FindNSRecordName 根据主键查找名称
func (this *NSRecordDAO) FindNSRecordName(tx *dbs.Tx, id uint64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}
