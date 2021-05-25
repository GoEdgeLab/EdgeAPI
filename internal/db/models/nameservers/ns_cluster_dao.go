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

// CreateCluster 创建集群
func (this *NSClusterDAO) CreateCluster(tx *dbs.Tx, name string) (int64, error) {
	op := NewNSClusterOperator()
	op.Name = name
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

// ListEnabledNSClusters 列出单页集群
func (this *NSClusterDAO) ListEnabledNSClusters(tx *dbs.Tx, offset int64, size int64) (result []*NSCluster, err error) {
	_, err = this.Query(tx).
		State(NSClusterStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledNSClusters 列出所有集群
func (this *NSClusterDAO) FindAllEnabledNSClusters(tx *dbs.Tx) (result []*NSCluster, err error) {
	_, err = this.Query(tx).
		State(NSClusterStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
