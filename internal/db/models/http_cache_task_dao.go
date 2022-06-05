package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

const (
	HTTPCacheTaskStateEnabled  = 1 // 已启用
	HTTPCacheTaskStateDisabled = 0 // 已禁用
)

type HTTPCacheTaskType = string

const (
	HTTPCacheTaskTypePurge HTTPCacheTaskType = "purge"
	HTTPCacheTaskTypeFetch HTTPCacheTaskType = "fetch"
)

type HTTPCacheTaskDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedHTTPCacheTaskDAO.Clean(nil, 30) // 只保留N天
				if err != nil {
					remotelogs.Error("HTTPCacheTaskDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

func NewHTTPCacheTaskDAO() *HTTPCacheTaskDAO {
	return dbs.NewDAO(&HTTPCacheTaskDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPCacheTasks",
			Model:  new(HTTPCacheTask),
			PkName: "id",
		},
	}).(*HTTPCacheTaskDAO)
}

var SharedHTTPCacheTaskDAO *HTTPCacheTaskDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPCacheTaskDAO = NewHTTPCacheTaskDAO()
	})
}

// EnableHTTPCacheTask 启用条目
func (this *HTTPCacheTaskDAO) EnableHTTPCacheTask(tx *dbs.Tx, taskId int64) error {
	_, err := this.Query(tx).
		Pk(taskId).
		Set("state", HTTPCacheTaskStateEnabled).
		Update()
	return err
}

// DisableHTTPCacheTask 禁用条目
func (this *HTTPCacheTaskDAO) DisableHTTPCacheTask(tx *dbs.Tx, taskId int64) error {
	_, err := this.Query(tx).
		Pk(taskId).
		Set("state", HTTPCacheTaskStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyChange(tx, taskId)
}

// FindEnabledHTTPCacheTask 查找启用中的条目
func (this *HTTPCacheTaskDAO) FindEnabledHTTPCacheTask(tx *dbs.Tx, taskId int64) (*HTTPCacheTask, error) {
	result, err := this.Query(tx).
		Pk(taskId).
		Attr("state", HTTPCacheTaskStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPCacheTask), err
}

// CreateTask 创建任务
func (this *HTTPCacheTaskDAO) CreateTask(tx *dbs.Tx, userId int64, taskType HTTPCacheTaskType, keyType string, description string) (int64, error) {
	var op = NewHTTPCacheTaskOperator()
	op.UserId = userId
	op.Type = taskType
	op.KeyType = keyType
	op.IsOk = false
	op.IsDone = false
	op.IsReady = false
	op.Description = description
	op.Day = timeutil.Format("Ymd")
	op.State = HTTPCacheTaskStateEnabled
	taskId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = this.NotifyChange(tx, taskId)
	if err != nil {
		return 0, err
	}
	return taskId, nil
}

// ResetTask 重置服务状态
func (this *HTTPCacheTaskDAO) ResetTask(tx *dbs.Tx, taskId int64) error {
	if taskId <= 0 {
		return errors.New("invalid 'taskId'")
	}

	var op = NewHTTPCacheTaskOperator()
	op.Id = taskId
	op.IsOk = false
	op.IsDone = false
	op.DoneAt = 0
	return this.Save(tx, op)
}

// UpdateTaskReady 设置任务为已准备
func (this *HTTPCacheTaskDAO) UpdateTaskReady(tx *dbs.Tx, taskId int64) error {
	return this.Query(tx).
		Pk(taskId).
		Set("isReady", true).
		UpdateQuickly()
}

// CountTasks 查询所有任务数量
func (this *HTTPCacheTaskDAO) CountTasks(tx *dbs.Tx, userId int64) (int64, error) {
	var query = this.Query(tx).
		State(HTTPCacheTaskStateEnabled).
		Attr("isReady", true)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	return query.Count()
}

// CountDoingTasks 查询正在执行的任务数量
func (this *HTTPCacheTaskDAO) CountDoingTasks(tx *dbs.Tx, userId int64) (int64, error) {
	var query = this.Query(tx).
		State(HTTPCacheTaskStateEnabled).
		Attr("isReady", true).
		Attr("isDone", false)
	if userId > 0 {
		query.Attr("userId", userId)
	}

	return query.Count()
}

// ListTasks 列出单页任务
func (this *HTTPCacheTaskDAO) ListTasks(tx *dbs.Tx, userId int64, offset int64, size int64) (result []*HTTPCacheTask, err error) {
	var query = this.Query(tx).
		State(HTTPCacheTaskStateEnabled).
		Attr("isReady", true)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	_, err = query.
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// ListDoingTasks 列出需要执行的任务
func (this *HTTPCacheTaskDAO) ListDoingTasks(tx *dbs.Tx, size int64) (result []*HTTPCacheTask, err error) {
	_, err = this.Query(tx).
		State(HTTPCacheTaskStateEnabled).
		Attr("isDone", false).
		Attr("isReady", true).
		Limit(size).
		AscPk(). // 按照先创建先执行的原则
		Slice(&result).
		FindAll()
	return
}

// UpdateTaskStatus 标记任务已完成
func (this *HTTPCacheTaskDAO) UpdateTaskStatus(tx *dbs.Tx, taskId int64, isDone bool, isOk bool) error {
	if taskId <= 0 {
		return errors.New("invalid taskId '" + types.String(taskId) + "'")
	}
	var op = NewHTTPCacheTaskOperator()
	op.Id = taskId
	op.IsDone = isDone
	op.IsOk = isOk

	if isDone {
		op.DoneAt = time.Now().Unix()
	}

	return this.Save(tx, op)
}

// CheckUserTask 检查用户任务
func (this *HTTPCacheTaskDAO) CheckUserTask(tx *dbs.Tx, userId int64, taskId int64) error {
	b, err := this.Query(tx).
		Pk(taskId).
		Attr("userId", userId).
		State(HTTPCacheTaskStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !b {
		return ErrNotFound
	}
	return nil
}

// Clean 清理以往的任务
func (this *HTTPCacheTaskDAO) Clean(tx *dbs.Tx, days int) error {
	if days <= 0 {
		days = 30
	}
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))

	// 删除Key
	err := SharedHTTPCacheTaskKeyDAO.Clean(tx, days)
	if err != nil {
		return err
	}

	// 删除任务
	_, err = this.Query(tx).
		Lte("day", day).
		Delete()
	return err
}

// NotifyChange 发送通知
func (this *HTTPCacheTaskDAO) NotifyChange(tx *dbs.Tx, taskId int64) error {
	// TODO
	return nil
}
