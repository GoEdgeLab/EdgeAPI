package metrics

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type MetricStatDAO dbs.DAO

func NewMetricStatDAO() *MetricStatDAO {
	return dbs.NewDAO(&MetricStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMetricStats",
			Model:  new(MetricStat),
			PkName: "id",
		},
	}).(*MetricStatDAO)
}

var SharedMetricStatDAO *MetricStatDAO

func init() {
	dbs.OnReady(func() {
		SharedMetricStatDAO = NewMetricStatDAO()
	})
}
