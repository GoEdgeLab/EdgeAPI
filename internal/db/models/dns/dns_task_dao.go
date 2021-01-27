package dns

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"time"
)

type DNSTaskType = string

const (
	DNSTaskTypeClusterChange DNSTaskType = "clusterChange"
	DNSTaskTypeNodeChange    DNSTaskType = "nodeChange"
	DNSTaskTypeServerChange  DNSTaskType = "serverChange"
	DNSTaskTypeDomainChange  DNSTaskType = "domainChange"
)

type DNSTaskDAO dbs.DAO

func NewDNSTaskDAO() *DNSTaskDAO {
	return dbs.NewDAO(&DNSTaskDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeDNSTasks",
			Model:  new(DNSTask),
			PkName: "id",
		},
	}).(*DNSTaskDAO)
}

var SharedDNSTaskDAO *DNSTaskDAO

func init() {
	dbs.OnReady(func() {
		SharedDNSTaskDAO = NewDNSTaskDAO()
	})
}

// 生成任务
func (this *DNSTaskDAO) CreateDNSTask(tx *dbs.Tx, clusterId int64, serverId int64, nodeId int64, domainId int64, taskType string) error {
	if clusterId <= 0 && serverId <= 0 && nodeId <= 0 {
		return nil
	}
	err := this.Query(tx).InsertOrUpdateQuickly(maps.Map{
		"clusterId": clusterId,
		"serverId":  serverId,
		"nodeId":    nodeId,
		"domainId":  domainId,
		"updatedAt": time.Now().Unix(),
		"type":      taskType,
		"isDone":    false,
		"isOk":      false,
		"error":     "",
	}, maps.Map{
		"updatedAt": time.Now().Unix(),
		"isDone":    false,
		"isOk":      false,
		"error":     "",
	})
	return err
}

// 生成集群任务
func (this *DNSTaskDAO) CreateClusterTask(tx *dbs.Tx, clusterId int64, taskType DNSTaskType) error {
	return this.CreateDNSTask(tx, clusterId, 0, 0, 0, taskType)
}

// 生成节点任务
func (this *DNSTaskDAO) CreateNodeTask(tx *dbs.Tx, nodeId int64, taskType DNSTaskType) error {
	return this.CreateDNSTask(tx, 0, 0, nodeId, 0, taskType)
}

// 生成服务任务
func (this *DNSTaskDAO) CreateServerTask(tx *dbs.Tx, serverId int64, taskType DNSTaskType) error {
	return this.CreateDNSTask(tx, 0, serverId, 0, 0, taskType)
}

// 生成域名更新任务
func (this *DNSTaskDAO) CreateDomainTask(tx *dbs.Tx, domainId int64, taskType DNSTaskType) error {
	return this.CreateDNSTask(tx, 0, 0, 0, domainId, taskType)
}

// 查找所有正在执行的任务
func (this *DNSTaskDAO) FindAllDoingTasks(tx *dbs.Tx) (result []*DNSTask, err error) {
	_, err = this.Query(tx).
		Attr("isDone", 0).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 查找正在执行的和错误的任务
func (this *DNSTaskDAO) FindAllDoingOrErrorTasks(tx *dbs.Tx) (result []*DNSTask, err error) {
	_, err = this.Query(tx).
		Where("(isDone=0 OR (isDone=1 AND isOk=0))").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 检查是否有正在执行的任务
func (this *DNSTaskDAO) ExistDoingTasks(tx *dbs.Tx) (bool, error) {
	return this.Query(tx).
		Attr("isDone", 0).
		Exist()
}

// 检查是否有错误的任务
func (this *DNSTaskDAO) ExistErrorTasks(tx *dbs.Tx) (bool, error) {
	return this.Query(tx).
		Attr("isDone", 1).
		Attr("isOk", 0).
		Exist()
}

// 删除任务
func (this *DNSTaskDAO) DeleteDNSTask(tx *dbs.Tx, taskId int64) error {
	_, err := this.Query(tx).
		Pk(taskId).
		Delete()
	return err
}

// 设置任务错误
func (this *DNSTaskDAO) UpdateDNSTaskError(tx *dbs.Tx, taskId int64, err string) error {
	if taskId <= 0 {
		return errors.New("invalid taskId")
	}
	op := NewDNSTaskOperator()
	op.Id = taskId
	op.IsDone = true
	op.Error = err
	op.IsOk = false
	return this.Save(tx, op)
}

// 设置任务完成
func (this *DNSTaskDAO) UpdateDNSTaskDone(tx *dbs.Tx, taskId int64) error {
	if taskId <= 0 {
		return errors.New("invalid taskId")
	}
	op := NewDNSTaskOperator()
	op.Id = taskId
	op.IsDone = true
	op.IsOk = true
	op.Error = ""
	return this.Save(tx, op)
}
