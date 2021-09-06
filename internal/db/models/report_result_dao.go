package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"time"
)

type ReportResultDAO dbs.DAO

func NewReportResultDAO() *ReportResultDAO {
	return dbs.NewDAO(&ReportResultDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeReportResults",
			Model:  new(ReportResult),
			PkName: "id",
		},
	}).(*ReportResultDAO)
}

var SharedReportResultDAO *ReportResultDAO

func init() {
	dbs.OnReady(func() {
		SharedReportResultDAO = NewReportResultDAO()
	})
}

// UpdateResult 创建结果
func (this *ReportResultDAO) UpdateResult(tx *dbs.Tx, taskType string, targetId int64, targetDesc string, reportNodeId int64, isOk bool, costMs float64, errString string) error {
	var countUp interface{} = 0
	var countDown interface{} = 0
	if isOk {
		countUp = dbs.SQL("countUp+1")
	} else {
		countDown = dbs.SQL("countDown+1")
	}

	return this.Query(tx).
		InsertOrUpdateQuickly(maps.Map{
			"type":         taskType,
			"targetId":     targetId,
			"targetDesc":   targetDesc,
			"updatedAt":    time.Now().Unix(),
			"reportNodeId": reportNodeId,
			"isOk":         isOk,
			"costMs":       costMs,
			"error":        errString,
			"countUp":      countUp,
			"countDown":    countDown,
		}, maps.Map{
			"targetDesc": targetDesc,
			"updatedAt":  time.Now().Unix(),
			"isOk":       isOk,
			"costMs":     costMs,
			"error":      errString,
			"countUp":    countUp,
			"countDown":  countDown,
		})
}

// CountAllResults 计算结果数量
func (this *ReportResultDAO) CountAllResults(tx *dbs.Tx, reportNodeId int64, okState configutils.BoolState) (int64, error) {
	var query = this.Query(tx).
		Attr("reportNodeId", reportNodeId)
	switch okState {
	case configutils.BoolStateYes:
		query.Attr("isOk", 1)
	case configutils.BoolStateNo:
		query.Attr("isOk", 0)
	}
	return query.
		Count()
}

// ListResults 列出单页结果
func (this *ReportResultDAO) ListResults(tx *dbs.Tx, reportNodeId int64, okState configutils.BoolState, offset int64, size int64) (result []*ReportResult, err error) {
	var query = this.Query(tx).
		Attr("reportNodeId", reportNodeId)
	switch okState {
	case configutils.BoolStateYes:
		query.Attr("isOk", 1)
	case configutils.BoolStateNo:
		query.Attr("isOk", 0)
	}
	_, err = query.
		Attr("reportNodeId", reportNodeId).
		Offset(offset).
		Limit(size).
		Desc("targetId").
		Slice(&result).
		FindAll()
	return
}
