package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strings"
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
	var day = timeutil.FormatTime("Ymd", createdAt)
	var hour = timeutil.FormatTime("YmdH", createdAt)
	var minute = timeutil.FormatTime("YmdHi", createdAt)

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

// Clean 清除数据
func (this *NodeValueDAO) Clean(tx *dbs.Tx) error {
	var hour = timeutil.Format("YmdH", time.Now().Add(-2*time.Hour))
	_, err := this.Query(tx).
		Where("hour<=:hour").
		Param("hour", hour).
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

	if err != nil {
		return nil, err
	}

	for _, nodeValue := range result {
		nodeValue.Value, _ = json.Marshal(maps.Map{
			key: types.Float32(string(nodeValue.Value)),
		})
	}

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

	if err != nil {
		return nil, err
	}

	for _, nodeValue := range result {
		nodeValue.Value, _ = json.Marshal(maps.Map{
			key: types.Float32(string(nodeValue.Value)),
		})
	}

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

	if err != nil {
		return nil, err
	}

	for _, nodeValue := range result {
		nodeValue.Value, _ = json.Marshal(maps.Map{
			key: types.Float32(string(nodeValue.Value)),
		})
	}

	return
}

// SumAllNodeValues 计算所有节点的某项参数值
func (this *NodeValueDAO) SumAllNodeValues(tx *dbs.Tx, role string, item nodeconfigs.NodeValueItem, param string, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit) (total float64, avg float64, max float64, err error) {
	if duration <= 0 {
		return 0, 0, 0, nil
	}

	var query = this.Query(tx).
		Result("SUM(JSON_EXTRACT(value, '$."+param+"')) AS sumValue", "AVG(JSON_EXTRACT(value, '$."+param+"')) AS avgValue", "MAX(JSON_EXTRACT(value, '$."+param+"')) AS maxValueResult"). // maxValue 是个MySQL Keyword，这里使用maxValueResult代替
		Attr("role", role).
		Attr("item", item)

	switch durationUnit {
	case nodeconfigs.NodeValueDurationUnitMinute:
		fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-int64(duration*60))
		query.Attr("minute", fromMinute)
	default:
		fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-int64(duration*60))
		query.Attr("minute", fromMinute)
	}

	m, _, err := query.FindOne()
	if err != nil {
		return 0, 0, 0, err
	}

	return m.GetFloat64("sumValue"), m.GetFloat64("avgValue"), m.GetFloat64("maxValueResult"), nil
}

// SumNodeValues 计算节点的某项参数值
func (this *NodeValueDAO) SumNodeValues(tx *dbs.Tx, role string, nodeId int64, item string, param string, method nodeconfigs.NodeValueSumMethod, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit) (float64, error) {
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

// SumNodeGroupValues 计算节点分组的某项参数值
func (this *NodeValueDAO) SumNodeGroupValues(tx *dbs.Tx, role string, groupId int64, item string, param string, method nodeconfigs.NodeValueSumMethod, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit) (float64, error) {
	if duration <= 0 {
		return 0, nil
	}

	query := this.Query(tx).
		Attr("role", role).
		Where("nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE groupId=:groupId AND state=1)").
		Param("groupId", groupId).
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

// SumNodeClusterValues 计算节点集群的某项参数值
func (this *NodeValueDAO) SumNodeClusterValues(tx *dbs.Tx, role string, clusterId int64, item string, param string, method nodeconfigs.NodeValueSumMethod, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit) (float64, error) {
	if duration <= 0 {
		return 0, nil
	}

	query := this.Query(tx).
		Attr("role", role).
		Attr("clusterId", clusterId).
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
	fromMinute := timeutil.FormatTime("YmdHi", time.Now().Unix()-int64(60))

	one, err := this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Attr("item", item).
		Attr("minute", fromMinute).
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*NodeValue), nil
}

// ComposeNodeStatus 组合节点状态值
func (this *NodeValueDAO) ComposeNodeStatus(tx *dbs.Tx, role string, nodeId int64, statusConfig *nodeconfigs.NodeStatus) error {
	var items = []string{
		nodeconfigs.NodeValueItemCPU,
		nodeconfigs.NodeValueItemMemory,
		nodeconfigs.NodeValueItemLoad,
		nodeconfigs.NodeValueItemTrafficOut,
		nodeconfigs.NodeValueItemTrafficIn,
	}
	ones, err := this.Query(tx).
		Result("item", "value").
		Attr("role", role).
		Attr("nodeId", nodeId).
		Attr("minute", timeutil.FormatTime("YmdHi", time.Now().Unix()-60)).
		Where("item IN ('" + strings.Join(items, "', '") + "')").
		FindAll()
	if err != nil {
		return err
	}
	for _, one := range ones {
		var oneValue = one.(*NodeValue)
		var valueMap = oneValue.DecodeMapValue()
		switch oneValue.Item {
		case nodeconfigs.NodeValueItemCPU:
			statusConfig.CPUUsage = valueMap.GetFloat64("usage")
		case nodeconfigs.NodeValueItemMemory:
			statusConfig.MemoryUsage = valueMap.GetFloat64("usage")
		case nodeconfigs.NodeValueItemLoad:
			statusConfig.Load1m = valueMap.GetFloat64("load1m")
			statusConfig.Load5m = valueMap.GetFloat64("load5m")
			statusConfig.Load15m = valueMap.GetFloat64("load15m")
		case nodeconfigs.NodeValueItemTrafficOut:
			statusConfig.TrafficOutBytes = valueMap.GetUint64("total")
		case nodeconfigs.NodeValueItemTrafficIn:
			statusConfig.TrafficInBytes = valueMap.GetUint64("total")
		}
	}

	return nil
}

// ComposeNodeStatusJSON 组合节点状态值，并转换为JSON数据
func (this *NodeValueDAO) ComposeNodeStatusJSON(tx *dbs.Tx, role string, nodeId int64, statusJSON []byte) ([]byte, error) {
	var statusConfig = &nodeconfigs.NodeStatus{}
	if len(statusJSON) > 0 {
		err := json.Unmarshal(statusJSON, statusConfig)
		if err != nil {
			return nil, err
		}
	}

	err := this.ComposeNodeStatus(tx, role, nodeId, statusConfig)
	if err != nil {
		return nil, err
	}

	return json.Marshal(statusConfig)
}
