package models

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

type APIMethodStatDAO dbs.DAO

func NewAPIMethodStatDAO() *APIMethodStatDAO {
	return dbs.NewDAO(&APIMethodStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeAPIMethodStats",
			Model:  new(APIMethodStat),
			PkName: "id",
		},
	}).(*APIMethodStatDAO)
}

var SharedAPIMethodStatDAO *APIMethodStatDAO

func init() {
	dbs.OnReady(func() {
		SharedAPIMethodStatDAO = NewAPIMethodStatDAO()
	})
}

// CreateStat 记录统计数据
func (this *APIMethodStatDAO) CreateStat(tx *dbs.Tx, method string, tag string, costMs float64) error {
	var day = timeutil.Format("Ymd")
	return this.Query(tx).
		Param("costMs", costMs).
		InsertOrUpdateQuickly(map[string]interface{}{
			"apiNodeId":  teaconst.NodeId,
			"method":     method,
			"tag":        tag,
			"costMs":     costMs,
			"peekMs":     costMs,
			"countCalls": 1,
			"day":        day,
		}, map[string]interface{}{
			"costMs":     dbs.SQL("(costMs*countCalls+:costMs)/(countCalls+1)"),
			"peekMs":     dbs.SQL("IF(peekMs>:costMs, peekMs, :costMs)"),
			"countCalls": dbs.SQL("countCalls+1"),
		})
}

// FindAllStatsWithDay 查询当前统计
func (this *APIMethodStatDAO) FindAllStatsWithDay(tx *dbs.Tx, day string) (result []*APIMethodStat, err error) {
	_, err = this.Query(tx).
		Attr("day", day).
		Slice(&result).
		FindAll()
	return
}

// CountAllStatsWithDay 统计当天数量
func (this *APIMethodStatDAO) CountAllStatsWithDay(tx *dbs.Tx, day string) (int64, error) {
	return this.Query(tx).
		Attr("day", day).
		Count()
}

// Clean 清理数据
func (this *APIMethodStatDAO) Clean(tx *dbs.Tx) error {
	var day = timeutil.Format("Ymd")
	_, err := this.Query(tx).
		Param("day", day).
		Where("day<:day").
		Delete()
	return err
}
