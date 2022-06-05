package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type HTTPCacheTaskKeyDAO dbs.DAO

func NewHTTPCacheTaskKeyDAO() *HTTPCacheTaskKeyDAO {
	return dbs.NewDAO(&HTTPCacheTaskKeyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPCacheTaskKeys",
			Model:  new(HTTPCacheTaskKey),
			PkName: "id",
		},
	}).(*HTTPCacheTaskKeyDAO)
}

var SharedHTTPCacheTaskKeyDAO *HTTPCacheTaskKeyDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPCacheTaskKeyDAO = NewHTTPCacheTaskKeyDAO()
	})
}

// CreateKey 创建Key
// 参数：
//   - clusterId 集群ID
//   - nodeMapJSON 集群下节点映射，格式类似于 `{ "节点1":true, ... }`
func (this *HTTPCacheTaskKeyDAO) CreateKey(tx *dbs.Tx, taskId int64, key string, taskType HTTPCacheTaskType, keyType string, clusterId int64) (int64, error) {
	var op = NewHTTPCacheTaskKeyOperator()
	op.TaskId = taskId
	op.Key = key
	op.Type = taskType
	op.KeyType = keyType
	op.ClusterId = clusterId

	op.Nodes = "{}"
	op.Errors = "{}"

	return this.SaveInt64(tx, op)
}

// UpdateKeyStatus 修改Key状态
func (this *HTTPCacheTaskKeyDAO) UpdateKeyStatus(tx *dbs.Tx, keyId int64, nodeId int64, errString string, nodesJSON []byte) error {
	if keyId <= 0 {
		return errors.New("invalid 'keyId'")
	}

	if len(nodesJSON) == 0 {
		nodesJSON = []byte("{}")
	}

	taskId, err := this.Query(tx).
		Pk(keyId).
		Result("taskId").
		FindInt64Col(0)
	if err != nil {
		return err
	}

	var jsonPath = "$.\"" + types.String(nodeId) + "\""

	var query = this.Query(tx).
		Pk(keyId).
		Set("nodes", dbs.SQL("JSON_SET(nodes, :jsonPath1, true)")).
		Param("jsonPath1", jsonPath)

	if len(errString) > 0 {
		query.Set("errors", dbs.SQL("JSON_SET(errors, :jsonPath2, :jsonValue2)")).
			Param("jsonPath2", jsonPath).
			Param("jsonValue2", errString)
	} else {
		query.Set("errors", dbs.SQL("JSON_REMOVE(errors, :jsonPath2)")).
			Param("jsonPath2", jsonPath)
	}

	err = query.
		UpdateQuickly()
	if err != nil {
		return err
	}

	// 检查是否已完成
	isDone, err := this.Query(tx).
		Pk(keyId).
		Where("JSON_CONTAINS(nodes, :nodesJSON)").
		Param("nodesJSON", nodesJSON).
		Exist()
	if err != nil {
		return err
	}

	if isDone {
		err = this.Query(tx).
			Pk(keyId).
			Set("isDone", isDone).
			UpdateQuickly()
		if err != nil {
			return err
		}

		// 检查任务是否已经完成
		taskIsNotDone, err := this.Query(tx).
			Attr("taskId", taskId).
			Attr("isDone", false).
			Exist()
		if err != nil {
			return err
		}
		var taskIsDone = !taskIsNotDone
		var hasErrors = true
		if taskIsDone {
			// 已经完成，是否有错误
			hasErrors, err = this.Query(tx).
				Attr("taskId", taskId).
				Where("JSON_LENGTH(errors)>0").
				Exist()
			if err != nil {
				return err
			}
		}
		err = SharedHTTPCacheTaskDAO.UpdateTaskStatus(tx, taskId, taskIsDone, !hasErrors)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindAllTaskKeys 查询某个任务下的所有Key
func (this *HTTPCacheTaskKeyDAO) FindAllTaskKeys(tx *dbs.Tx, taskId int64) (result []*HTTPCacheTaskKey, err error) {
	_, err = this.Query(tx).
		Attr("taskId", taskId).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindDoingTaskKeys 查询要执行的任务
func (this *HTTPCacheTaskKeyDAO) FindDoingTaskKeys(tx *dbs.Tx, nodeId int64, size int64) (result []*HTTPCacheTaskKey, err error) {
	// 集群ID
	clusterIds, err := SharedNodeDAO.FindEnabledAndOnNodeClusterIds(tx, nodeId)
	if err != nil {
		return nil, err
	}

	if len(clusterIds) == 0 {
		return nil, nil
	}

	_, err = this.Query(tx).
		Attr("clusterId", clusterIds).
		Attr("isDone", false).
		Where("NOT JSON_CONTAINS_PATH(nodes, 'one', :jsonPath1)").
		Param("jsonPath1", "$.\""+types.String(nodeId)+"\"").
		Where("taskId IN (SELECT id FROM " + SharedHTTPCacheTaskDAO.Table + " WHERE state=1 AND isReady=1 AND isDone=0)").
		Limit(size).
		AscPk().
		Reuse(false).
		Slice(&result).
		FindAll()
	if err != nil {
		return nil, err
	}

	return
}

// ResetCacheKeysWithTaskId 重置任务下的Key状态
func (this *HTTPCacheTaskKeyDAO) ResetCacheKeysWithTaskId(tx *dbs.Tx, taskId int64) error {
	return this.Query(tx).
		Attr("taskId", taskId).
		Set("isDone", false).
		Set("nodes", "{}").
		Set("errors", "{}").
		UpdateQuickly()
}

// CountUserTasksInDay 读取某个用户当前数量
// day YYYYMMDD
func (this *HTTPCacheTaskKeyDAO) CountUserTasksInDay(tx *dbs.Tx, userId int64, day string, taskType HTTPCacheTaskType) (int64, error) {
	if userId <= 0 {
		return 0, nil
	}

	// 这里需要包含已删除的
	return this.Query(tx).
		Where("taskId IN (SELECT id FROM "+SharedHTTPCacheTaskDAO.Table+" WHERE userId=:userId AND day=:day AND type=:type)").
		Param("userId", userId).
		Param("day", day).
		Param("type", taskType).
		Count()
}

// Clean 清理以往的任务
func (this *HTTPCacheTaskKeyDAO) Clean(tx *dbs.Tx, days int) error {
	if days <= 0 {
		days = 30
	}

	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Where("taskId IN (SELECT id FROM "+SharedHTTPCacheTaskDAO.Table+" WHERE day<=:day)").
		Param("day", day).
		Delete()
	return err
}
