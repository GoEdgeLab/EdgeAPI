//go:build !plus

package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSRouteStateEnabled  = 1 // 已启用
	NSRouteStateDisabled = 0 // 已禁用
)

type NSRouteDAO dbs.DAO

func NewNSRouteDAO() *NSRouteDAO {
	return dbs.NewDAO(&NSRouteDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSRoutes",
			Model:  new(NSRoute),
			PkName: "id",
		},
	}).(*NSRouteDAO)
}

var SharedNSRouteDAO *NSRouteDAO

func init() {
	dbs.OnReady(func() {
		SharedNSRouteDAO = NewNSRouteDAO()
	})
}

// EnableNSRoute 启用条目
func (this *NSRouteDAO) EnableNSRoute(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSRouteStateEnabled).
		Update()
	return err
}

// DisableNSRoute 禁用条目
func (this *NSRouteDAO) DisableNSRoute(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSRouteStateDisabled).
		Update()
	return err
}

// FindEnabledNSRoute 查找启用中的条目
func (this *NSRouteDAO) FindEnabledNSRoute(tx *dbs.Tx, id uint32) (*NSRoute, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(NSRouteStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSRoute), err
}

// FindNSRouteName 根据主键查找名称
func (this *NSRouteDAO) FindNSRouteName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}
