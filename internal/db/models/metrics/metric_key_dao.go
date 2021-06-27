package metrics

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type MetricKeyDAO dbs.DAO

func NewMetricKeyDAO() *MetricKeyDAO {
	return dbs.NewDAO(&MetricKeyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMetricKeys",
			Model:  new(MetricKey),
			PkName: "id",
		},
	}).(*MetricKeyDAO)
}

var SharedMetricKeyDAO *MetricKeyDAO

func init() {
	dbs.OnReady(func() {
		SharedMetricKeyDAO = NewMetricKeyDAO()
	})
}
