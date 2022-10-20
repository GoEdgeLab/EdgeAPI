package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	TrafficPackagePeriodStateEnabled  = 1 // 已启用
	TrafficPackagePeriodStateDisabled = 0 // 已禁用
)

type TrafficPackagePeriodDAO dbs.DAO

func NewTrafficPackagePeriodDAO() *TrafficPackagePeriodDAO {
	return dbs.NewDAO(&TrafficPackagePeriodDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeTrafficPackagePeriods",
			Model:  new(TrafficPackagePeriod),
			PkName: "id",
		},
	}).(*TrafficPackagePeriodDAO)
}

var SharedTrafficPackagePeriodDAO *TrafficPackagePeriodDAO

func init() {
	dbs.OnReady(func() {
		SharedTrafficPackagePeriodDAO = NewTrafficPackagePeriodDAO()
	})
}

// EnableTrafficPackagePeriod 启用条目
func (this *TrafficPackagePeriodDAO) EnableTrafficPackagePeriod(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", TrafficPackagePeriodStateEnabled).
		Update()
	return err
}

// DisableTrafficPackagePeriod 禁用条目
func (this *TrafficPackagePeriodDAO) DisableTrafficPackagePeriod(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", TrafficPackagePeriodStateDisabled).
		Update()
	return err
}

// FindEnabledTrafficPackagePeriod 查找启用中的条目
func (this *TrafficPackagePeriodDAO) FindEnabledTrafficPackagePeriod(tx *dbs.Tx, id int64) (*TrafficPackagePeriod, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(TrafficPackagePeriodStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*TrafficPackagePeriod), err
}
