package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSZoneStateEnabled  = 1 // 已启用
	NSZoneStateDisabled = 0 // 已禁用
)

type NSZoneDAO dbs.DAO

func NewNSZoneDAO() *NSZoneDAO {
	return dbs.NewDAO(&NSZoneDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSZones",
			Model:  new(NSZone),
			PkName: "id",
		},
	}).(*NSZoneDAO)
}

var SharedNSZoneDAO *NSZoneDAO

func init() {
	dbs.OnReady(func() {
		SharedNSZoneDAO = NewNSZoneDAO()
	})
}

// EnableNSZone 启用条目
func (this *NSZoneDAO) EnableNSZone(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSZoneStateEnabled).
		Update()
	return err
}

// DisableNSZone 禁用条目
func (this *NSZoneDAO) DisableNSZone(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSZoneStateDisabled).
		Update()
	return err
}

// FindEnabledNSZone 查找启用中的条目
func (this *NSZoneDAO) FindEnabledNSZone(tx *dbs.Tx, id uint64) (*NSZone, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSZoneStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSZone), err
}
