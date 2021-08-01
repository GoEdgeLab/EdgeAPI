package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type NodeValueDAO dbs.DAO

func NewNodeValueDAO() *NodeValueDAO {
	return dbs.NewDAO(&NodeValueDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeValues",
			Model:  new(NodeValue),
			PkName: "id",
		},
	}).(*NodeValueDAO)
}

var SharedNodeValueDAO *NodeValueDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeValueDAO = NewNodeValueDAO()
	})
}

// CreateValue 创建值
func (this *NodeValueDAO) CreateValue(tx *dbs.Tx, clusterId int64, role nodeconfigs.NodeRole, nodeId int64, item string, valueJSON []byte, createdAt int64) error {
	day := timeutil.FormatTime("Ymd", createdAt)
	hour := timeutil.FormatTime("YmdH", createdAt)
	minute := timeutil.FormatTime("YmdHi", createdAt)

	return this.Query(tx).
		InsertOrUpdateQuickly(maps.Map{
			"clusterId": clusterId,
			"role":      role,
			"nodeId":    nodeId,
			"item":      item,
			"value":     valueJSON,
			"createdAt": createdAt,
			"day":       day,
			"hour":      hour,
			"minute":    minute,
		}, maps.Map{
			"value": valueJSON,
		})
}

// DeleteExpiredValues 清除数据
func (this *NodeValueDAO) DeleteExpiredValues(tx *dbs.Tx) error {
	// 删除N天之前的所有数据
	expiredDays := 100
	day := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -expiredDays))
	_, err := this.Query(tx).
		Where("day<:day").
		Param("day", day).
		Delete()
	if err != nil {
		return err
	}
	return nil
}

// ListValues 列出最近的的数据
func (this *NodeValueDAO) ListValues(tx *dbs.Tx, role string, nodeId int64, item string, timeRange nodeconfigs.NodeValueRange) (result []*NodeValue, err error) {
	query := this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Attr("item", item)

	switch timeRange {
	// TODO 支持更多的时间范围
	case nodeconfigs.NodeValueRangeMinute:
		fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-3600) // 一个小时之前的
		query.Gte("minute", fromMinute)
	default:
		err = errors.New("invalid 'range' value: '" + timeRange + "'")
		return
	}

	_, err = query.Slice(&result).
		FindAll()
	return
}

// ListValuesWithClusterId 列出集群最近的的平均数据
func (this *NodeValueDAO) ListValuesWithClusterId(tx *dbs.Tx, role string, clusterId int64, item string, key string, timeRange nodeconfigs.NodeValueRange) (result []*NodeValue, err error) {
	query := this.Query(tx).
		Attr("role", role).
		Attr("item", item).
		Result("AVG(JSON_EXTRACT(value, '$." + key + "')) AS value, MIN(createdAt) AS createdAt")

	switch role {
	case nodeconfigs.NodeRoleNode:
		query.Where("nodeId IN (SELECT id FROM " + SharedNodeDAO.Table + " WHERE (clusterId=:clusterId OR JSON_CONTAINS(secondaryClusterIds, :clusterIdString)) AND state=1)")
		query.Param("clusterId", clusterId).
			Param("clusterIdString", types.String(clusterId))
	default:
		query.Attr("clusterId", clusterId)
	}

	switch timeRange {
	// TODO 支持更多的时间范围
	case nodeconfigs.NodeValueRangeMinute:
		fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-3600) // 一个小时之前的
		query.Gte("minute", fromMinute)
		query.Result("minute")
		query.Group("minute")
	default:
		err = errors.New("invalid 'range' value: '" + timeRange + "'")
		return
	}

	_, err = query.Slice(&result).
		FindAll()
	return
}

// ListValuesForUserNodes 列出用户节点相关的平均数据
func (this *NodeValueDAO) ListValuesForUserNodes(tx *dbs.Tx, item string, key string, timeRange nodeconfigs.NodeValueRange) (result []*NodeValue, err error) {
	query := this.Query(tx).
		Attr("role", "user").
		Attr("item", item).
		Result("AVG(JSON_EXTRACT(value, '$." + key + "')) AS value, MIN(createdAt) AS createdAt")

	switch timeRange {
	// TODO 支持更多的时间范围
	case nodeconfigs.NodeValueRangeMinute:
		fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-3600) // 一个小时之前的
		query.Gte("minute", fromMinute)
		query.Result("minute")
		query.Group("minute")
	default:
		err = errors.New("invalid 'range' value: '" + timeRange + "'")
		return
	}

	_, err = query.Slice(&result).
		FindAll()
	return
}

// ListValuesForNSNodes 列出用户节点相关的平均数据
func (this *NodeValueDAO) ListValuesForNSNodes(tx *dbs.Tx, item string, key string, timeRange nodeconfigs.NodeValueRange) (result []*NodeValue, err error) {
	query := this.Query(tx).
		Attr("role", "dns").
		Attr("item", item).
		Result("AVG(JSON_EXTRACT(value, '$." + key + "')) AS value, MIN(createdAt) AS createdAt")

	switch timeRange {
	// TODO 支持更多的时间范围
	case nodeconfigs.NodeValueRangeMinute:
		fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-3600) // 一个小时之前的
		query.Gte("minute", fromMinute)
		query.Result("minute")
		query.Group("minute")
	default:
		err = errors.New("invalid 'range' value: '" + timeRange + "'")
		return
	}

	_, err = query.Slice(&result).
		FindAll()
	return
}

// SumValues 计算某项参数值
func (this *NodeValueDAO) SumValues(tx *dbs.Tx, role string, nodeId int64, item string, param string, method nodeconfigs.NodeValueSumMethod, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit) (float64, error) {
	if duration <= 0 {
		return 0, nil
	}

	query := this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Attr("item", item)
	switch method {
	case nodeconfigs.NodeValueSumMethodAvg:
		query.Result("AVG(JSON_EXTRACT(value, '$." + param + "'))")
	case nodeconfigs.NodeValueSumMethodSum:
		query.Result("SUM(JSON_EXTRACT(value, '$." + param + "'))")
	default:
		query.Result("AVG(JSON_EXTRACT(value, '$." + param + "'))")
	}
	switch durationUnit {
	case nodeconfigs.NodeValueDurationUnitMinute:
		fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-int64(duration*60))
		query.Gte("minute", fromMinute)
	default:
		fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-int64(duration*60))
		query.Gte("minute", fromMinute)
	}
	return query.FindFloat64Col(0)
}

// FindLatestNodeValue 获取最近一条数据
func (this *NodeValueDAO) FindLatestNodeValue(tx *dbs.Tx, role string, nodeId int64, item string) (*NodeValue, error) {
	one, err := this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Attr("item", item).
		DescPk().
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*NodeValue), nil
}
