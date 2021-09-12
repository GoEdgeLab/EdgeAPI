package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/reporterconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
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
func (this *ReportResultDAO) UpdateResult(tx *dbs.Tx, taskType string, targetId int64, targetDesc string, reportNodeId int64, level reporterconfigs.ReportLevel, isOk bool, costMs float64, errString string) error {
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
			"level":        level,
		}, maps.Map{
			"targetDesc": targetDesc,
			"updatedAt":  time.Now().Unix(),
			"isOk":       isOk,
			"costMs":     costMs,
			"error":      errString,
			"countUp":    countUp,
			"countDown":  countDown,
			"level":      level,
		})
}

// CountAllResults 计算结果数量
func (this *ReportResultDAO) CountAllResults(tx *dbs.Tx, reportNodeId int64, level reporterconfigs.ReportLevel, okState configutils.BoolState) (int64, error) {
	var query = this.Query(tx).
		Attr("reportNodeId", reportNodeId)
	switch okState {
	case configutils.BoolStateYes:
		query.Attr("isOk", 1)
	case configutils.BoolStateNo:
		query.Attr("isOk", 0)
	}
	if len(level) > 0 {
		query.Attr("level", level)
	}
	return query.
		Gt("updatedAt", time.Now().Unix()-600).
		Count()
}

// ListResults 列出单页结果
func (this *ReportResultDAO) ListResults(tx *dbs.Tx, reportNodeId int64, okState configutils.BoolState, level reporterconfigs.ReportLevel, offset int64, size int64) (result []*ReportResult, err error) {
	var query = this.Query(tx).
		Attr("reportNodeId", reportNodeId)
	switch okState {
	case configutils.BoolStateYes:
		query.Attr("isOk", 1)
	case configutils.BoolStateNo:
		query.Attr("isOk", 0)
	}
	if len(level) > 0 {
		query.Attr("level", level)
	}
	_, err = query.
		Attr("reportNodeId", reportNodeId).
		Gt("updatedAt", time.Now().Unix()-600).
		Offset(offset).
		Limit(size).
		Desc("targetId").
		Slice(&result).
		FindAll()
	return
}

// FindAvgCostMsWithTarget 获取某个对象的平均耗时
func (this *ReportResultDAO) FindAvgCostMsWithTarget(tx *dbs.Tx, taskType reporterconfigs.TaskType, targetId int64) (float64, error) {
	return this.Query(tx).
		Attr("type", taskType).
		Attr("targetId", targetId).
		Where("reportNodeId IN (SELECT id FROM "+SharedReportNodeDAO.Table+" WHERE state=1 AND isOn=1)").
		Attr("isOk", true).
		Gt("updatedAt", time.Now().Unix()-600).
		Avg("costMs", 0)
}

// FindAvgLevelWithTarget 获取某个对象的平均级别
func (this *ReportResultDAO) FindAvgLevelWithTarget(tx *dbs.Tx, taskType reporterconfigs.TaskType, targetId int64) (string, error) {
	ones, _, err := this.Query(tx).
		Result("COUNT(*) AS c, level").
		Attr("type", taskType).
		Attr("targetId", targetId).
		Where("reportNodeId IN (SELECT id FROM "+SharedReportNodeDAO.Table+" WHERE state=1 AND isOn=1)").
		Gt("updatedAt", time.Now().Unix()-600).
		Group("level").
		FindOnes()
	if err != nil {
		return "", err
	}

	if len(ones) == 0 {
		return reporterconfigs.ReportLevelNormal, nil
	}

	var total = 0
	var levelMap = map[string]int{} // code => count
	for _, one := range ones {
		var c = one.GetInt("c")
		total += c
		levelMap[one.GetString("level")] = c
	}
	if total == 0 {
		return reporterconfigs.ReportLevelNormal, nil
	}

	var half = total / 2
	for _, def := range reporterconfigs.FindAllReportLevels() {
		c, ok := levelMap[def.Code]
		if ok {
			half -= c
			if half <= 0 {
				return def.Code, nil
			}
		}
	}

	return "", nil
}

// FindConnectivityWithTarget 获取某个对象的连通率
func (this *ReportResultDAO) FindConnectivityWithTarget(tx *dbs.Tx, taskType reporterconfigs.TaskType, targetId int64, groupId int64) (float64, error) {
	var query = this.Query(tx).
		Attr("type", taskType).
		Attr("targetId", targetId)
	if groupId > 0 {
		query.Where("reportNodeId IN (SELECT id FROM "+SharedReportNodeDAO.Table+" WHERE state=1 AND isOn=1 AND JSON_CONTAINS(groupIds, :groupIdString))").
			Param("groupIdString", types.String(groupId))
	} else {
		query.Where("reportNodeId IN (SELECT id FROM " + SharedReportNodeDAO.Table + " WHERE state=1 AND isOn=1)")
	}

	// 已汇报数据的数量
	total, err := query.
		Gt("updatedAt", time.Now().Unix()-600).
		Count()
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 1, nil
	}

	// 连通的数量
	var connectedQuery = this.Query(tx).
		Attr("type", taskType).
		Attr("targetId", targetId)

	if groupId > 0 {
		connectedQuery.Where("reportNodeId IN (SELECT id FROM "+SharedReportNodeDAO.Table+" WHERE state=1 AND isOn=1 AND JSON_CONTAINS(groupIds, :groupIdString))").
			Param("groupIdString", types.String(groupId))
	} else {
		connectedQuery.Where("reportNodeId IN (SELECT id FROM " + SharedReportNodeDAO.Table + " WHERE state=1 AND isOn=1)")
	}

	countConnected, err := connectedQuery.
		Attr("isOk", true).
		Gt("updatedAt", time.Now().Unix()-600).
		Count()
	if err != nil {
		return 0, err
	}
	return float64(countConnected) / float64(total), nil
}
