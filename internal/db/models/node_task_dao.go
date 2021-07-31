package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"time"
)

type NodeTaskType = string

const (
	NodeTaskTypeConfigChanged      NodeTaskType = "configChanged"
	NodeTaskTypeIPItemChanged      NodeTaskType = "ipItemChanged"
	NodeTaskTypeNodeVersionChanged NodeTaskType = "nodeVersionChanged"
)

type NodeTaskDAO dbs.DAO

func NewNodeTaskDAO() *NodeTaskDAO {
	return dbs.NewDAO(&NodeTaskDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeTasks",
			Model:  new(NodeTask),
			PkName: "id",
		},
	}).(*NodeTaskDAO)
}

var SharedNodeTaskDAO *NodeTaskDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeTaskDAO = NewNodeTaskDAO()
	})
}

// CreateNodeTask 创建单个节点任务
func (this *NodeTaskDAO) CreateNodeTask(tx *dbs.Tx, clusterId int64, nodeId int64, taskType NodeTaskType) error {
	if clusterId <= 0 || nodeId <= 0 {
		return nil
	}
	uniqueId := numberutils.FormatInt64(nodeId) + "@node@" + taskType
	updatedAt := time.Now().Unix()
	_, _, err := this.Query(tx).
		InsertOrUpdate(maps.Map{
			"clusterId": clusterId,
			"nodeId":    nodeId,
			"type":      taskType,
			"uniqueId":  uniqueId,
			"updatedAt": updatedAt,
			"isDone":    0,
			"isOk":      0,
			"error":     "",
		}, maps.Map{
			"clusterId": clusterId,
			"updatedAt": updatedAt,
			"isDone":    0,
			"isOk":      0,
			"error":     "",
		})
	return err
}

// CreateClusterTask 创建集群任务
func (this *NodeTaskDAO) CreateClusterTask(tx *dbs.Tx, clusterId int64, taskType NodeTaskType) error {
	if clusterId <= 0 {
		return nil
	}

	uniqueId := numberutils.FormatInt64(clusterId) + "@cluster@" + taskType
	updatedAt := time.Now().Unix()
	_, _, err := this.Query(tx).
		InsertOrUpdate(maps.Map{
			"clusterId":  clusterId,
			"nodeId":     0,
			"type":       taskType,
			"uniqueId":   uniqueId,
			"updatedAt":  updatedAt,
			"isDone":     0,
			"isOk":       0,
			"isNotified": 0,
			"error":      "",
		}, maps.Map{
			"updatedAt":  updatedAt,
			"isDone":     0,
			"isOk":       0,
			"isNotified": 0,
			"error":      "",
		})
	return err
}

// ExtractClusterTask 分解集群任务
func (this *NodeTaskDAO) ExtractClusterTask(tx *dbs.Tx, clusterId int64, taskType NodeTaskType) error {
	nodeIds, err := SharedNodeDAO.FindAllNodeIdsMatch(tx, clusterId, true, configutils.BoolStateYes)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Attr("clusterId", clusterId).
		Param("clusterIdString", types.String(clusterId)).
		Where("nodeId> 0").
		Attr("type", taskType).
		Delete()
	if err != nil {
		return err
	}

	for _, nodeId := range nodeIds {
		err = this.CreateNodeTask(tx, clusterId, nodeId, taskType)
		if err != nil {
			return err
		}
	}

	_, err = this.Query(tx).
		Attr("clusterId", clusterId).
		Attr("nodeId", 0).
		Attr("type", taskType).
		Delete()
	if err != nil {
		return err
	}

	return nil
}

// ExtractAllClusterTasks 分解所有集群任务
func (this *NodeTaskDAO) ExtractAllClusterTasks(tx *dbs.Tx) error {
	ones, err := this.Query(tx).
		Attr("nodeId", 0).
		FindAll()
	if err != nil {
		return err
	}
	for _, one := range ones {
		clusterId := int64(one.(*NodeTask).ClusterId)
		err = this.ExtractClusterTask(tx, clusterId, one.(*NodeTask).Type)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteAllClusterTasks 删除集群所有相关任务
func (this *NodeTaskDAO) DeleteAllClusterTasks(tx *dbs.Tx, clusterId int64) error {
	_, err := this.Query(tx).
		Attr("clusterId", clusterId).
		Delete()
	return err
}

// DeleteNodeTasks 删除节点相关任务
func (this *NodeTaskDAO) DeleteNodeTasks(tx *dbs.Tx, nodeId int64) error {
	_, err := this.Query(tx).
		Attr("nodeId", nodeId).
		Delete()
	return err
}

// FindDoingNodeTasks 查询一个节点的所有任务
func (this *NodeTaskDAO) FindDoingNodeTasks(tx *dbs.Tx, nodeId int64) (result []*NodeTask, err error) {
	if nodeId <= 0 {
		return
	}
	_, err = this.Query(tx).
		Attr("nodeId", nodeId).
		Where("(isDone=0 OR (isDone=1 AND isOk=0))").
		Slice(&result).
		FindAll()
	return
}

// UpdateNodeTaskDone 修改节点任务的完成状态
func (this *NodeTaskDAO) UpdateNodeTaskDone(tx *dbs.Tx, taskId int64, isOk bool, errorMessage string) error {
	_, err := this.Query(tx).
		Pk(taskId).
		Set("isDone", 1).
		Set("isOk", isOk).
		Set("error", errorMessage).
		Update()
	return err
}

// FindAllDoingTaskClusterIds 查找正在更新的集群IDs
func (this *NodeTaskDAO) FindAllDoingTaskClusterIds(tx *dbs.Tx) ([]int64, error) {
	ones, _, err := this.Query(tx).
		Result("DISTINCT(clusterId) AS clusterId").
		Where("(nodeId=0 OR (isDone=0 OR (isDone=1 AND isOk=0)))").
		FindOnes()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, one.GetInt64("clusterId"))
	}
	return result, nil
}

// FindAllDoingNodeTasksWithClusterId 查询某个集群下所有的任务
func (this *NodeTaskDAO) FindAllDoingNodeTasksWithClusterId(tx *dbs.Tx, clusterId int64) (result []*NodeTask, err error) {
	_, err = this.Query(tx).
		Attr("clusterId", clusterId).
		Gt("nodeId", 0).
		Where("(isDone=0 OR (isDone=1 AND isOk=0))").
		Desc("isDone").
		Asc().
		Asc("nodeId").
		Slice(&result).
		FindAll()
	return
}

// ExistsDoingNodeTasks 检查是否有正在执行的任务
func (this *NodeTaskDAO) ExistsDoingNodeTasks(tx *dbs.Tx) (bool, error) {
	return this.Query(tx).
		Where("(isDone=0 OR (isDone=1 AND isOk=0))").
		Gt("nodeId", 0).
		Exist()
}

// ExistsErrorNodeTasks 是否有错误的任务
func (this *NodeTaskDAO) ExistsErrorNodeTasks(tx *dbs.Tx) (bool, error) {
	return this.Query(tx).
		Where("(isDone=1 AND isOk=0)").
		Exist()
}

// DeleteNodeTask 删除任务
func (this *NodeTaskDAO) DeleteNodeTask(tx *dbs.Tx, taskId int64) error {
	_, err := this.Query(tx).
		Pk(taskId).
		Delete()
	return err
}

// CountDoingNodeTasks 计算正在执行的任务
func (this *NodeTaskDAO) CountDoingNodeTasks(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		Attr("isDone", 0).
		Gt("nodeId", 0).
		Count()
}

// FindNotifyingNodeTasks 查找需要通知的任务
func (this *NodeTaskDAO) FindNotifyingNodeTasks(tx *dbs.Tx, size int64) (result []*NodeTask, err error) {
	_, err = this.Query(tx).
		Gt("nodeId", 0).
		Attr("isNotified", 0).
		Attr("isDone", 0).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// UpdateTasksNotified 设置任务已通知
func (this *NodeTaskDAO) UpdateTasksNotified(tx *dbs.Tx, taskIds []int64) error {
	if len(taskIds) == 0 {
		return nil
	}
	for _, taskId := range taskIds {
		_, err := this.Query(tx).
			Pk(taskId).
			Set("isNotified", 1).
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}
