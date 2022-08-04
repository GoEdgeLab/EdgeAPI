package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
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

// CreateCluster 创建集群
func (this *NSClusterDAO) CreateCluster(tx *dbs.Tx, name string, accessLogRefJSON []byte) (int64, error) {
	var op = NewNSClusterOperator()
	op.Name = name

	if len(accessLogRefJSON) > 0 {
		op.AccessLog = accessLogRefJSON
	}

	op.IsOn = true
	op.State = NSClusterStateEnabled

	// 默认端口
	// TCP
	{
		var config = &serverconfigs.TCPProtocolConfig{}
		config.IsOn = true
		config.Listen = []*serverconfigs.NetworkAddressConfig{
			{
				Protocol:  serverconfigs.ProtocolTCP,
				PortRange: "53",
			},
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			return 0, err
		}
		op.Tcp = configJSON
	}

	// UDP
	{
		var config = &serverconfigs.UDPProtocolConfig{}
		config.IsOn = true
		config.Listen = []*serverconfigs.NetworkAddressConfig{
			{
				Protocol:  serverconfigs.ProtocolUDP,
				PortRange: "53",
			},
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			return 0, err
		}
		op.Udp = configJSON
	}

	return this.SaveInt64(tx, op)
}

// UpdateCluster 修改集群
func (this *NSClusterDAO) UpdateCluster(tx *dbs.Tx, clusterId int64, name string, isOn bool) error {
	if clusterId <= 0 {
		return errors.New("invalid clusterId")
	}
	var op = NewNSClusterOperator()
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

// FindClusterGrantId 查找集群的认证ID
func (this *NSClusterDAO) FindClusterGrantId(tx *dbs.Tx, clusterId int64) (int64, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("grantId").
		FindInt64Col(0)
}

// UpdateRecursion 设置递归DNS
func (this *NSClusterDAO) UpdateRecursion(tx *dbs.Tx, clusterId int64, recursionJSON []byte) error {
	err := this.Query(tx).
		Pk(clusterId).
		Set("recursion", recursionJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterRecursion 读取递归DNS配置
func (this *NSClusterDAO) FindClusterRecursion(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	recursion, err := this.Query(tx).
		Result("recursion").
		Pk(clusterId).
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	return []byte(recursion), nil
}

// FindClusterTCP 查找集群的TCP设置
func (this *NSClusterDAO) FindClusterTCP(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("tcp").
		FindBytesCol()
}

// UpdateClusterTCP 修改集群的TCP设置
func (this *NSClusterDAO) UpdateClusterTCP(tx *dbs.Tx, clusterId int64, tcpConfig *serverconfigs.TCPProtocolConfig) error {
	tcpJSON, err := json.Marshal(tcpConfig)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("tcp", tcpJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterTLS 查找集群的TLS设置
func (this *NSClusterDAO) FindClusterTLS(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("tls").
		FindBytesCol()
}

// UpdateClusterTLS 修改集群的TLS设置
func (this *NSClusterDAO) UpdateClusterTLS(tx *dbs.Tx, clusterId int64, tlsConfig *serverconfigs.TLSProtocolConfig) error {
	tlsJSON, err := json.Marshal(tlsConfig)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("tls", tlsJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
}

// FindClusterUDP 查找集群的TCP设置
func (this *NSClusterDAO) FindClusterUDP(tx *dbs.Tx, clusterId int64) ([]byte, error) {
	return this.Query(tx).
		Pk(clusterId).
		Result("udp").
		FindBytesCol()
}

// UpdateClusterUDP 修改集群的UDP设置
func (this *NSClusterDAO) UpdateClusterUDP(tx *dbs.Tx, clusterId int64, udpConfig *serverconfigs.UDPProtocolConfig) error {
	udpJSON, err := json.Marshal(udpConfig)
	if err != nil {
		return err
	}
	err = this.Query(tx).
		Pk(clusterId).
		Set("udp", udpJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, clusterId)
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
