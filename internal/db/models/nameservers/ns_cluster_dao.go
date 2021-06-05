package nameservers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSClusterStateEnabled  = 1 // 已启用
	NSClusterStateDisabled = 0 // 已禁用
)

type NSClusterDAO dbs.DAO

func NewNSClusterDAO() *NSClusterDAO {
	return dbs.NewDAO(&NSClusterDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSClusters",
			Model:  new(NSCluster),
			PkName: "id",
		},
	}).(*NSClusterDAO)
}

var SharedNSClusterDAO *NSClusterDAO

func init() {
	dbs.OnReady(func() {
		SharedNSClusterDAO = NewNSClusterDAO()
	})
}

// EnableNSCluster 启用条目
func (this *NSClusterDAO) EnableNSCluster(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSClusterStateEnabled).
		Update()
	return err
}

// DisableNSCluster 禁用条目
func (this *NSClusterDAO) DisableNSCluster(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSClusterStateDisabled).
		Update()
	return err
}

// FindEnabledNSCluster 查找启用中的条目
func (this *NSClusterDAO) FindEnabledNSCluster(tx *dbs.Tx, id int64) (*NSCluster, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSClusterStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSCluster), err
}

// FindEnabledNSClusterName 查找启用中的条目名称
func (this *NSClusterDAO) FindEnabledNSClusterName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		State(NSClusterStateEnabled).
		Result("name").
		FindStringCol("")
}

// CreateCluster 创建集群
func (this *NSClusterDAO) CreateCluster(tx *dbs.Tx, name string, accessLogRefJSON []byte) (int64, error) {
	op := NewNSClusterOperator()
	op.Name = name

	if len(accessLogRefJSON) > 0 {
		op.AccessLog = accessLogRefJSON
	}

	op.IsOn = true
	op.State = NSClusterStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateCluster 修改集群
func (this *NSClusterDAO) UpdateCluster(tx *dbs.Tx, clusterId int64, name string, isOn bool) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	op := NewNSClusterOperator()
	op.Id = clusterId
	op.Name = name
	op.IsOn = isOn
	return this.Save(tx, op)
}

// CountAllEnabledClusters 计算可用集群数量
func (this *NSClusterDAO) CountAllEnabledClusters(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NSClusterStateEnabled).
		Count()
}

// ListEnabledClusters 列出单页集群
func (this *NSClusterDAO) ListEnabledClusters(tx *dbs.Tx, offset int64, size int64) (result []*NSCluster, err error) {
	_, err = this.Query(tx).
		State(NSClusterStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledClusters 列出所有集群
func (this *NSClusterDAO) FindAllEnabledClusters(tx *dbs.Tx) (result []*NSCluster, err error) {
	_, err = this.Query(tx).
		State(NSClusterStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// UpdateClusterAccessLog 设置访问日志
func (this *NSClusterDAO) UpdateClusterAccessLog(tx *dbs.Tx, clusterId int64, accessLogJSON []byte) error {
	return this.Query(tx).
		Pk(clusterId).
		Set("accessLog", accessLogJSON).
		UpdateQuickly()
}

// FindClusterAccessLog 读取访问日志配置
func (this *NSClusterDAO) FindClusterAccessLog(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	accessLog, err := this.Query(tx).
		Pk(clusterId).
		Result("accessLog").
		FindStringCol("")
	return []byte(accessLog), err
}
