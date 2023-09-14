package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"strings"
	"time"
)

type NodeTaskType = string

const (
	// CDN相关

	NodeTaskTypeConfigChanged             NodeTaskType = "configChanged"             // 节点整体配置变化
	NodeTaskTypeDDosProtectionChanged     NodeTaskType = "ddosProtectionChanged"     // 节点DDoS配置变更
	NodeTaskTypeGlobalServerConfigChanged NodeTaskType = "globalServerConfigChanged" // 全局服务设置变化
	NodeTaskTypeIPListDeleted             NodeTaskType = "ipListDeleted"             // IPList被删除
	NodeTaskTypeIPItemChanged             NodeTaskType = "ipItemChanged"             // IP条目变更
	NodeTaskTypeNodeVersionChanged        NodeTaskType = "nodeVersionChanged"        // 节点版本变化
	NodeTaskTypeScriptsChanged            NodeTaskType = "scriptsChanged"            // 脚本配置变化
	NodeTaskTypeNodeLevelChanged          NodeTaskType = "nodeLevelChanged"          // 节点级别变化
	NodeTaskTypeUserServersStateChanged   NodeTaskType = "userServersStateChanged"   // 用户服务状态变化
	NodeTaskTypeUAMPolicyChanged          NodeTaskType = "uamPolicyChanged"          // UAM策略变化
	NodeTaskTypeHTTPPagesPolicyChanged    NodeTaskType = "httpPagesPolicyChanged"    // 自定义页面变化
	NodeTaskTypeHTTPCCPolicyChanged       NodeTaskType = "httpCCPolicyChanged"       // CC策略变化
	NodeTaskTypeHTTP3PolicyChanged        NodeTaskType = "http3PolicyChanged"        // HTTP3策略变化
	NodeTaskTypeUpdatingServers           NodeTaskType = "updatingServers"           // 更新一组服务
	NodeTaskTypeTOAChanged                NodeTaskType = "toaChanged"                // TOA配置变化

	// NS相关

	NSNodeTaskTypeConfigChanged         NodeTaskType = "nsConfigChanged"
	NSNodeTaskTypeDomainChanged         NodeTaskType = "nsDomainChanged"
	NSNodeTaskTypeRecordChanged         NodeTaskType = "nsRecordChanged"
	NSNodeTaskTypeRouteChanged          NodeTaskType = "nsRouteChanged"
	NSNodeTaskTypeKeyChanged            NodeTaskType = "nsKeyChanged"
	NSNodeTaskTypeDDosProtectionChanged NodeTaskType = "nsDDoSProtectionChanged" // 节点DDoS配置变更
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
func (this *NodeTaskDAO) CreateNodeTask(tx *dbs.Tx, role string, clusterId int64, nodeId int64, userId int64, serverId int64, taskType NodeTaskType) error {
	if clusterId <= 0 || nodeId <= 0 {
		return nil
	}
	var uniqueId = role + "@" + types.String(nodeId) + "@node@" + types.String(serverId) + "@" + taskType

	// 用户信息
	// 没有直接加入到 uniqueId 中，是为了兼容以前的字段值
	if userId > 0 {
		uniqueId += "@" + types.String(userId)
	}

	version, err := this.increaseVersion(tx)
	if err != nil {
		return err
	}

	var updatedAt = time.Now().Unix()
	_, _, err = this.Query(tx).
		InsertOrUpdate(maps.Map{
			"role":      role,
			"clusterId": clusterId,
			"nodeId":    nodeId,
			"userId":    userId,
			"serverId":  serverId,
			"type":      taskType,
			"uniqueId":  uniqueId,
			"updatedAt": updatedAt,
			"isDone":    0,
			"isOk":      0,
			"error":     "",
			"version":   version,
		}, maps.Map{
			"clusterId":  clusterId,
			"updatedAt":  updatedAt,
			"isDone":     0,
			"isOk":       0,
			"error":      "",
			"isNotified": 0,
			"version":    version,
			"serverId":   serverId,
		})
	return err
}

// CreateClusterTask 创建集群任务
func (this *NodeTaskDAO) CreateClusterTask(tx *dbs.Tx, role string, clusterId int64, userId int64, serverId int64, taskType NodeTaskType) error {
	if clusterId <= 0 {
		return nil
	}

	var uniqueId = role + "@" + types.String(clusterId) + "@" + types.String(serverId) + "@cluster@" + taskType

	// 用户信息
	// 没有直接加入到 uniqueId 中，是为了兼容以前的字段值
	if userId > 0 {
		uniqueId += "@" + types.String(userId)
	}

	var updatedAt = time.Now().Unix()
	_, _, err := this.Query(tx).
		InsertOrUpdate(maps.Map{
			"role":       role,
			"clusterId":  clusterId,
			"userId":     userId,
			"serverId":   serverId,
			"nodeId":     0,
			"type":       taskType,
			"uniqueId":   uniqueId,
			"updatedAt":  updatedAt,
			"isDone":     0,
			"isOk":       0,
			"isNotified": 0,
			"error":      "",
			"version":    time.Now().UnixNano(),
		}, maps.Map{
			"updatedAt":  updatedAt,
			"isDone":     0,
			"isOk":       0,
			"isNotified": 0,
			"error":      "",
			"version":    time.Now().UnixNano(),
			"serverId":   serverId,
		})
	return err
}

// ExtractNodeClusterTask 分解边缘节点集群任务
func (this *NodeTaskDAO) ExtractNodeClusterTask(tx *dbs.Tx, clusterId int64, userId int64, serverId int64, taskType NodeTaskType) error {
	nodeIds, err := SharedNodeDAO.FindAllNodeIdsMatch(tx, clusterId, true, configutils.BoolStateYes)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Attr("role", nodeconfigs.NodeRoleNode).
		Attr("clusterId", clusterId).
		Attr("serverId", serverId).
		Gt("nodeId", 0).
		Attr("type", taskType).
		Delete()
	if err != nil {
		return err
	}

	for _, nodeId := range nodeIds {
		err = this.CreateNodeTask(tx, nodeconfigs.NodeRoleNode, clusterId, nodeId, userId, serverId, taskType)
		if err != nil {
			return err
		}
	}

	_, err = this.Query(tx).
		Attr("role", nodeconfigs.NodeRoleNode).
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
func (this *NodeTaskDAO) ExtractAllClusterTasks(tx *dbs.Tx, role string) error {
	ones, err := this.Query(tx).
		Attr("role", role).
		Attr("nodeId", 0).
		FindAll()
	if err != nil {
		return err
	}
	for _, one := range ones {
		var clusterId = int64(one.(*NodeTask).ClusterId)
		switch role {
		case nodeconfigs.NodeRoleNode:
			var nodeTask = one.(*NodeTask)
			err = this.ExtractNodeClusterTask(tx, clusterId, int64(nodeTask.UserId), int64(nodeTask.ServerId), nodeTask.Type)
			if err != nil {
				return err
			}
		case nodeconfigs.NodeRoleDNS:
			err = this.ExtractNSClusterTask(tx, clusterId, one.(*NodeTask).Type)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteAllClusterTasks 删除集群所有相关任务
func (this *NodeTaskDAO) DeleteAllClusterTasks(tx *dbs.Tx, role string, clusterId int64) error {
	_, err := this.Query(tx).
		Attr("role", role).
		Attr("clusterId", clusterId).
		Delete()
	return err
}

// DeleteNodeTasks 删除节点相关任务
func (this *NodeTaskDAO) DeleteNodeTasks(tx *dbs.Tx, role string, nodeId int64) error {
	_, err := this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Delete()
	return err
}

// DeleteAllNodeTasks 删除所有节点相关任务
func (this *NodeTaskDAO) DeleteAllNodeTasks(tx *dbs.Tx) error {
	return this.Query(tx).
		DeleteQuickly()
}

// FindDoingNodeTasks 查询一个节点的所有任务
func (this *NodeTaskDAO) FindDoingNodeTasks(tx *dbs.Tx, role string, nodeId int64, version int64) (result []*NodeTask, err error) {
	if nodeId <= 0 {
		return
	}
	var query = this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		UseIndex("nodeId").
		Asc("version")
	if version > 0 {
		query.Lt("LENGTH(version)", 19) // 兼容以往版本
		query.Gt("version", version)
	} else {
		// 第一次访问时只取当前正在执行的或者执行失败的
		query.Where("(isDone=0 OR (isDone=1 AND isOk=0))")
	}
	_, err = query.
		Slice(&result).
		FindAll()
	return
}

// UpdateNodeTaskDone 修改节点任务的完成状态
func (this *NodeTaskDAO) UpdateNodeTaskDone(tx *dbs.Tx, taskId int64, isOk bool, errorMessage string) error {
	if isOk {
		// 特殊任务删除
		taskType, err := this.Query(tx).
			Pk(taskId).
			Result("type").
			FindStringCol("")
		if err != nil {
			return err
		}
		if strings.HasPrefix(taskType, NodeTaskTypeIPListDeleted+"@") {
			return this.Query(tx).
				Pk(taskId).
				DeleteQuickly()
		}
	}

	// 其他任务标记为完成
	var query = this.Query(tx).
		Pk(taskId)
	if !isOk {
		version, err := this.increaseVersion(tx)
		if err != nil {
			return err
		}
		query.Set("version", version)
	}

	_, err := query.
		Set("isDone", true).
		Set("isOk", isOk).
		Set("error", errorMessage).
		Update()
	return err
}

// FindAllDoingTaskClusterIds 查找正在更新的集群IDs
func (this *NodeTaskDAO) FindAllDoingTaskClusterIds(tx *dbs.Tx, role string) ([]int64, error) {
	ones, _, err := this.Query(tx).
		Result("DISTINCT(clusterId) AS clusterId").
		Attr("role", role).
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
func (this *NodeTaskDAO) FindAllDoingNodeTasksWithClusterId(tx *dbs.Tx, role string, clusterId int64) (result []*NodeTask, err error) {
	_, err = this.Query(tx).
		Attr("role", role).
		Attr("clusterId", clusterId).
		Gt("nodeId", 0).
		Where("(isDone=0 OR (isDone=1 AND isOk=0))").
		Desc("isDone").
		Asc("nodeId").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllDoingNodeIds 查询有任务的节点IDs
func (this *NodeTaskDAO) FindAllDoingNodeIds(tx *dbs.Tx, role string) ([]int64, error) {
	ones, err := this.Query(tx).
		Result("DISTINCT(nodeId) AS nodeId").
		Attr("role", role).
		Gt("nodeId", 0).
		Attr("isDone", false).
		Attr("isNotified", 0).
		FindAll()
	if err != nil {
		return nil, err
	}
	var result []int64
	for _, one := range ones {
		result = append(result, int64(one.(*NodeTask).NodeId))
	}
	return result, nil
}

// ExistsDoingNodeTasks 检查是否有正在执行的任务
func (this *NodeTaskDAO) ExistsDoingNodeTasks(tx *dbs.Tx, role string, excludeTypes []NodeTaskType) (bool, error) {
	var query = this.Query(tx).
		Attr("role", role).
		Where("(isDone=0 OR (isDone=1 AND isOk=0))").
		Gt("nodeId", 0)
	if len(excludeTypes) > 0 {
		for _, excludeType := range excludeTypes {
			query.Neq("type", excludeType)
		}
	}
	return query.Exist()
}

// ExistsErrorNodeTasks 是否有错误的任务
func (this *NodeTaskDAO) ExistsErrorNodeTasks(tx *dbs.Tx, role string, excludeTypes []NodeTaskType) (bool, error) {
	var query = this.Query(tx).
		Attr("role", role).
		Where("(isDone=1 AND isOk=0)")
	if len(excludeTypes) > 0 {
		for _, excludeType := range excludeTypes {
			query.Neq("type", excludeType)
		}
	}
	return query.Exist()
}

// DeleteNodeTask 删除任务
func (this *NodeTaskDAO) DeleteNodeTask(tx *dbs.Tx, taskId int64) error {
	_, err := this.Query(tx).
		Pk(taskId).
		Delete()
	return err
}

// CountDoingNodeTasks 计算正在执行的任务
func (this *NodeTaskDAO) CountDoingNodeTasks(tx *dbs.Tx, role string) (int64, error) {
	return this.Query(tx).
		Attr("isDone", 0).
		Attr("role", role).
		Gt("nodeId", 0).
		Count()
}

// FindNotifyingNodeTasks 查找需要通知的任务
func (this *NodeTaskDAO) FindNotifyingNodeTasks(tx *dbs.Tx, role string, size int64) (result []*NodeTask, err error) {
	_, err = this.Query(tx).
		Attr("role", role).
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

// 生成一个版本号
func (this *NodeTaskDAO) increaseVersion(tx *dbs.Tx) (version int64, err error) {
	return SharedSysLockerDAO.Increase(tx, "NODE_TASK_VERSION", 0)
}
