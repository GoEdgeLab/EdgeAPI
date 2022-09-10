package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"strconv"
	"strings"
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
func (this *NSClusterDAO) DisableNSCluster(tx *dbs.Tx, clusterId int64) error {
	_, err := this.Query(tx).
		Pk(clusterId).
		Set("state", NSClusterStateDisabled).
		Update()
	if err != nil {
		return err
	}

	return SharedNodeLogDAO.DeleteNodeLogsWithCluster(tx, nodeconfigs.NodeRoleDNS, clusterId)
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

// FindAllEnabledClusterIds 获取所有集群IDs
func (this *NSClusterDAO) FindAllEnabledClusterIds(tx *dbs.Tx) ([]int64, error) {
	ones, err := this.Query(tx).
		State(NSClusterStateEnabled).
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	var result = []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*NSCluster).Id))
	}
	return result, nil
}

// FindClusterGrantId 查找集群的认证ID
func (this *NSClusterDAO) FindClusterGrantId(tx *dbs.Tx, clusterId int64) (int64, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("grantId").
		FindInt64Col(0)
}

// CountAllClustersWithSSLPolicyIds 计算使用SSL策略的所有NS集群数量
func (this *NSClusterDAO) CountAllClustersWithSSLPolicyIds(tx *dbs.Tx, sslPolicyIds []int64) (count int64, err error) {
	if len(sslPolicyIds) == 0 {
		return
	}
	policyStringIds := []string{}
	for _, policyId := range sslPolicyIds {
		policyStringIds = append(policyStringIds, strconv.FormatInt(policyId, 10))
	}
	return this.Query(tx).
		State(NSClusterStateEnabled).
		Where("(FIND_IN_SET(JSON_EXTRACT(tls, '$.sslPolicyRef.sslPolicyId'), :policyIds)) ").
		Param("policyIds", strings.Join(policyStringIds, ",")).
		Count()
}

// NotifyUpdate 通知更改
func (this *NSClusterDAO) NotifyUpdate(tx *dbs.Tx, clusterId int64) error {
	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleDNS, clusterId, 0, NSNodeTaskTypeConfigChanged)
}
