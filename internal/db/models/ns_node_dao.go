package models

import (
	"encoding/json"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"time"
)

const (
	NSNodeStateEnabled  = 1 // 已启用
	NSNodeStateDisabled = 0 // 已禁用
)

type NSNodeDAO dbs.DAO

func NewNSNodeDAO() *NSNodeDAO {
	return dbs.NewDAO(&NSNodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSNodes",
			Model:  new(NSNode),
			PkName: "id",
		},
	}).(*NSNodeDAO)
}

var SharedNSNodeDAO *NSNodeDAO

func init() {
	dbs.OnReady(func() {
		SharedNSNodeDAO = NewNSNodeDAO()
	})
}

// EnableNSNode 启用条目
func (this *NSNodeDAO) EnableNSNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSNodeStateEnabled).
		Update()
	return err
}

// DisableNSNode 禁用条目
func (this *NSNodeDAO) DisableNSNode(tx *dbs.Tx, nodeId int64) error {
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("state", NSNodeStateDisabled).
		Update()

	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, nodeId)
	if err != nil {
		return err
	}

	// 删除运行日志
	return SharedNodeLogDAO.DeleteNodeLogs(tx, nodeconfigs.NodeRoleDNS, nodeId)
}

// FindEnabledNSNode 查找启用中的条目
func (this *NSNodeDAO) FindEnabledNSNode(tx *dbs.Tx, id int64) (*NSNode, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSNode), err
}

// FindEnabledNSNodeName 查找节点名称
func (this *NSNodeDAO) FindEnabledNSNodeName(tx *dbs.Tx, nodeId int64) (string, error) {
	return this.Query(tx).
		Pk(nodeId).
		State(NSNodeStateEnabled).
		Result("name").
		FindStringCol("")
}

// FindAllEnabledNodesWithClusterId 查找一个集群下的所有节点
func (this *NSNodeDAO) FindAllEnabledNodesWithClusterId(tx *dbs.Tx, clusterId int64) (result []*NSNode, err error) {
	_, err = this.Query(tx).
		Attr("clusterId", clusterId).
		State(NSNodeStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledNodes 所有集群的可用的节点数量
func (this *NSNodeDAO) CountAllEnabledNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NSNodeStateEnabled).
		Where("clusterId IN (SELECT id FROM " + SharedNSClusterDAO.Table + " WHERE state=1)").
		Count()
}

// CountAllOfflineNodes 计算离线节点数量
func (this *NSNodeDAO) CountAllOfflineNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NSNodeStateEnabled).
		Where("(status IS NULL OR JSON_EXTRACT(status, '$.updatedAt')<UNIX_TIMESTAMP()-120)").
		Where("clusterId IN (SELECT id FROM " + SharedNSClusterDAO.Table + " WHERE state=1)").
		Count()
}

// CountAllEnabledNodesMatch 计算满足条件的节点数量
func (this *NSNodeDAO) CountAllEnabledNodesMatch(tx *dbs.Tx, clusterId int64, installState configutils.BoolState, activeState configutils.BoolState, keyword string) (int64, error) {
	query := this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	// 安装状态
	switch installState {
	case configutils.BoolStateAll:
		// 所有
	case configutils.BoolStateYes:
		query.Attr("isInstalled", 1)
	case configutils.BoolStateNo:
		query.Attr("isInstalled", 0)
	}

	// 在线状态
	switch activeState {
	case configutils.BoolStateAll:
		// 所有
	case configutils.BoolStateYes:
		query.Where("(isActive=1 AND JSON_EXTRACT(status, '$.isActive') AND UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')<=60)")
	case configutils.BoolStateNo:
		query.Where("(isActive=0 OR status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>60)")
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}

	return query.
		State(NSNodeStateEnabled).
		Count()
}

// ListAllEnabledNodesMatch 列出单页匹配的节点
func (this *NSNodeDAO) ListAllEnabledNodesMatch(tx *dbs.Tx, clusterId int64, installState configutils.BoolState, activeState configutils.BoolState, keyword string, offset int64, size int64) (result []*NSNode, err error) {
	query := this.Query(tx)

	// 安装状态
	switch installState {
	case configutils.BoolStateAll:
		// 所有
	case configutils.BoolStateYes:
		query.Attr("isInstalled", 1)
	case configutils.BoolStateNo:
		query.Attr("isInstalled", 0)
	}

	// 在线状态
	switch activeState {
	case configutils.BoolStateAll:
		// 所有
	case configutils.BoolStateYes:
		query.Where("(isActive=1 AND JSON_EXTRACT(status, '$.isActive') AND UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')<=60)")
	case configutils.BoolStateNo:
		query.Where("(isActive=0 OR status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>60)")
	}

	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	_, err = query.
		State(NSNodeStateEnabled).
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// CountAllLowerVersionNodesWithClusterId 计算单个集群中所有低于某个版本的节点数量
func (this *NSNodeDAO) CountAllLowerVersionNodesWithClusterId(tx *dbs.Tx, clusterId int64, os string, arch string, version string) (int64, error) {
	return this.Query(tx).
		State(NSNodeStateEnabled).
		Attr("clusterId", clusterId).
		Where("status IS NOT NULL").
		Where("JSON_EXTRACT(status, '$.os')=:os").
		Where("JSON_EXTRACT(status, '$.arch')=:arch").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("os", os).
		Param("arch", arch).
		Param("version", utils.VersionToLong(version)).
		Count()
}

// CreateNode 创建节点
func (this *NSNodeDAO) CreateNode(tx *dbs.Tx, adminId int64, name string, clusterId int64) (nodeId int64, err error) {
	uniqueId, err := this.GenUniqueId(tx)
	if err != nil {
		return 0, err
	}

	secret := rands.String(32)

	// 保存API Token
	err = SharedApiTokenDAO.CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleDNS)
	if err != nil {
		return
	}

	var op = NewNSNodeOperator()
	op.AdminId = adminId
	op.Name = name
	op.UniqueId = uniqueId
	op.Secret = secret
	op.ClusterId = clusterId
	op.IsOn = 1
	op.State = NSNodeStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	// 通知节点更新
	nodeId = types.Int64(op.Id)
	err = this.NotifyUpdate(tx, nodeId)
	if err != nil {
		return 0, err
	}

	// 通知DNS更新
	err = this.NotifyDNSUpdate(tx, nodeId)
	if err != nil {
		return 0, err
	}

	return nodeId, nil
}

// UpdateNode 修改节点
func (this *NSNodeDAO) UpdateNode(tx *dbs.Tx, nodeId int64, name string, clusterId int64, isOn bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	var op = NewNSNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.ClusterId = clusterId
	op.IsOn = isOn
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, nodeId)
	if err != nil {
		return err
	}

	return this.NotifyDNSUpdate(tx, nodeId)
}

// FindEnabledNodeIdWithUniqueId 根据唯一ID获取节点ID
func (this *NSNodeDAO) FindEnabledNodeIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", NSNodeStateEnabled).
		ResultPk().
		FindInt64Col(0)
}

// FindNodeInstallStatus 查询节点的安装状态
func (this *NSNodeDAO) FindNodeInstallStatus(tx *dbs.Tx, nodeId int64) (*NodeInstallStatus, error) {
	node, err := this.Query(tx).
		Pk(nodeId).
		Result("installStatus", "isInstalled").
		Find()
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("not found")
	}

	installStatus := node.(*NSNode).InstallStatus
	isInstalled := node.(*NSNode).IsInstalled
	if len(installStatus) == 0 {
		return NewNodeInstallStatus(), nil
	}

	status := &NodeInstallStatus{}
	err = json.Unmarshal(installStatus, status)
	if err != nil {
		return nil, err
	}
	if isInstalled {
		status.IsFinished = true
		status.IsOk = true
	}
	return status, nil
}

// GenUniqueId 生成唯一ID
func (this *NSNodeDAO) GenUniqueId(tx *dbs.Tx) (string, error) {
	for {
		uniqueId := rands.HexString(32)
		ok, err := this.Query(tx).
			Attr("uniqueId", uniqueId).
			Exist()
		if err != nil {
			return "", err
		}
		if ok {
			continue
		}
		return uniqueId, nil
	}
}

// UpdateNodeIsInstalled 设置节点安装状态
func (this *NSNodeDAO) UpdateNodeIsInstalled(tx *dbs.Tx, nodeId int64, isInstalled bool) error {
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("isInstalled", isInstalled).
		Set("installStatus", "null"). // 重置安装状态
		Update()
	return err
}

// UpdateNodeStatus 更改节点状态
func (this *NSNodeDAO) UpdateNodeStatus(tx *dbs.Tx, nodeId int64, nodeStatus *nodeconfigs.NodeStatus) error {
	if nodeStatus == nil {
		return nil
	}

	nodeStatusJSON, err := json.Marshal(nodeStatus)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(nodeId).
		Set("status", nodeStatusJSON).
		Update()
	return err
}

// CountAllLowerVersionNodes 计算所有节点中低于某个版本的节点数量
func (this *NSNodeDAO) CountAllLowerVersionNodes(tx *dbs.Tx, version string) (int64, error) {
	return this.Query(tx).
		State(NSNodeStateEnabled).
		Where("clusterId IN (SELECT id FROM "+SharedNSClusterDAO.Table+" WHERE state=1)").
		Where("status IS NOT NULL").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("version", utils.VersionToLong(version)).
		Count()
}

// ComposeNodeConfig 组合节点配置
func (this *NSNodeDAO) ComposeNodeConfig(tx *dbs.Tx, nodeId int64) (*dnsconfigs.NSNodeConfig, error) {
	if nodeId <= 0 {
		return nil, nil
	}
	node, err := this.FindEnabledNSNode(tx, nodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	cluster, err := SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(node.ClusterId))
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, nil
	}

	var config = &dnsconfigs.NSNodeConfig{
		Id:        int64(node.Id),
		NodeId:    node.UniqueId,
		Secret:    node.Secret,
		ClusterId: int64(node.ClusterId),
	}

	// 访问日志
	// 全局配置
	{
		globalValue, err := SharedSysSettingDAO.ReadSetting(tx, systemconfigs.SettingCodeNSAccessLogSetting)
		if err != nil {
			return nil, err
		}
		if len(globalValue) > 0 {
			var ref = &dnsconfigs.NSAccessLogRef{}
			err = json.Unmarshal(globalValue, ref)
			if err != nil {
				return nil, err
			}
			config.AccessLogRef = ref
		}

		// 集群配置
		if len(cluster.AccessLog) > 0 {
			ref := &dnsconfigs.NSAccessLogRef{}
			err = json.Unmarshal(cluster.AccessLog, ref)
			if err != nil {
				return nil, err
			}
			if ref.IsPrior {
				config.AccessLogRef = ref
			}
		}
	}

	// 递归DNS配置
	if IsNotNull(cluster.Recursion) {
		var recursionConfig = &dnsconfigs.RecursionConfig{}
		err = json.Unmarshal(cluster.Recursion, recursionConfig)
		if err != nil {
			return nil, err
		}
		config.RecursionConfig = recursionConfig
	}

	// TCP
	if IsNotNull(cluster.Tcp) {
		var tcpConfig = &serverconfigs.TCPProtocolConfig{}
		err = json.Unmarshal(cluster.Tcp, tcpConfig)
		if err != nil {
			return nil, err
		}
		config.TCP = tcpConfig
	}

	// TLS
	if IsNotNull(cluster.Tls) {
		var tlsConfig = &serverconfigs.TLSProtocolConfig{}
		err = json.Unmarshal(cluster.Tls, tlsConfig)
		if err != nil {
			return nil, err
		}

		// SSL
		if tlsConfig.SSLPolicyRef != nil {
			sslPolicyConfig, err := SharedSSLPolicyDAO.ComposePolicyConfig(tx, tlsConfig.SSLPolicyRef.SSLPolicyId, nil)
			if err != nil {
				return nil, err
			}
			if sslPolicyConfig != nil {
				tlsConfig.SSLPolicy = sslPolicyConfig
			}
		}

		config.TLS = tlsConfig
	}

	// UDP
	if IsNotNull(cluster.Udp) {
		var udpConfig = &serverconfigs.UDPProtocolConfig{}
		err = json.Unmarshal(cluster.Udp, udpConfig)
		if err != nil {
			return nil, err
		}
		config.UDP = udpConfig
	}

	return config, nil
}

// FindNodeClusterId 获取节点的集群ID
func (this *NSNodeDAO) FindNodeClusterId(tx *dbs.Tx, nodeId int64) (int64, error) {
	return this.Query(tx).
		Pk(nodeId).
		Result("clusterId").
		FindInt64Col(0)
}

// FindNodeActive 检查节点活跃状态
func (this *NSNodeDAO) FindNodeActive(tx *dbs.Tx, nodeId int64) (bool, error) {
	isActive, err := this.Query(tx).
		Pk(nodeId).
		Result("isActive").
		FindIntCol(0)
	if err != nil {
		return false, err
	}
	return isActive == 1, nil
}

// UpdateNodeActive 修改节点活跃状态
func (this *NSNodeDAO) UpdateNodeActive(tx *dbs.Tx, nodeId int64, isActive bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("isActive", isActive).
		Set("statusIsNotified", false).
		Set("inactiveNotifiedAt", 0).
		Update()
	return err
}

// UpdateNodeConnectedAPINodes 修改当前连接的API节点
func (this *NSNodeDAO) UpdateNodeConnectedAPINodes(tx *dbs.Tx, nodeId int64, apiNodeIds []int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	var op = NewNSNodeOperator()
	op.Id = nodeId

	if len(apiNodeIds) > 0 {
		apiNodeIdsJSON, err := json.Marshal(apiNodeIds)
		if err != nil {
			return errors.Wrap(err)
		}
		op.ConnectedAPINodes = apiNodeIdsJSON
	} else {
		op.ConnectedAPINodes = "[]"
	}
	err := this.Save(tx, op)
	return err
}

// FindAllNotifyingInactiveNodesWithClusterId 取得某个集群所有等待通知离线离线的节点
func (this *NSNodeDAO) FindAllNotifyingInactiveNodesWithClusterId(tx *dbs.Tx, clusterId int64) (result []*NSNode, err error) {
	_, err = this.Query(tx).
		State(NSNodeStateEnabled).
		Attr("clusterId", clusterId).
		Attr("isOn", true).        // 只监控启用的节点
		Attr("isInstalled", true). // 只监控已经安装的节点
		Attr("isActive", false).   // 当前已经离线的
		Attr("statusIsNotified", false).
		Result("id", "name").
		Slice(&result).
		FindAll()
	return
}

// UpdateNodeStatusIsNotified 设置状态已经通知
func (this *NSNodeDAO) UpdateNodeStatusIsNotified(tx *dbs.Tx, nodeId int64) error {
	return this.Query(tx).
		Pk(nodeId).
		Set("statusIsNotified", true).
		Set("inactiveNotifiedAt", time.Now().Unix()).
		UpdateQuickly()
}

// FindNodeInactiveNotifiedAt 读取上次的节点离线通知时间
func (this *NSNodeDAO) FindNodeInactiveNotifiedAt(tx *dbs.Tx, nodeId int64) (int64, error) {
	return this.Query(tx).
		Pk(nodeId).
		Result("inactiveNotifiedAt").
		FindInt64Col(0)
}

// FindAllNodeIdsMatch 匹配节点并返回节点ID
func (this *NSNodeDAO) FindAllNodeIdsMatch(tx *dbs.Tx, clusterId int64, includeSecondaryNodes bool, isOn configutils.BoolState) (result []int64, err error) {
	query := this.Query(tx)
	query.State(NSNodeStateEnabled)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if isOn == configutils.BoolStateYes {
		query.Attr("isOn", true)
	} else if isOn == configutils.BoolStateNo {
		query.Attr("isOn", false)
	}
	query.Result("id")
	ones, _, err := query.FindOnes()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		result = append(result, one.GetInt64("id"))
	}
	return
}

// UpdateNodeInstallStatus 修改节点的安装状态
func (this *NSNodeDAO) UpdateNodeInstallStatus(tx *dbs.Tx, nodeId int64, status *NodeInstallStatus) error {
	if status == nil {
		_, err := this.Query(tx).
			Pk(nodeId).
			Set("installStatus", "null").
			Update()
		return err
	}

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	_, err = this.Query(tx).
		Pk(nodeId).
		Set("installStatus", string(data)).
		Update()
	return err
}

// FindEnabledNodeIdsWithClusterId 查找集群下的所有节点
func (this *NSNodeDAO) FindEnabledNodeIdsWithClusterId(tx *dbs.Tx, clusterId int64) ([]int64, error) {
	if clusterId <= 0 {
		return nil, nil
	}
	ones, err := this.Query(tx).
		ResultPk().
		Attr("clusterId", clusterId).
		State(NSNodeStateEnabled).
		FindAll()
	if err != nil {
		return nil, err
	}
	var result = []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*NSNode).Id))
	}
	return result, nil
}

// NotifyUpdate 通知更新
func (this *NSNodeDAO) NotifyUpdate(tx *dbs.Tx, nodeId int64) error {
	// TODO 先什么都不做
	return nil
}

// NotifyDNSUpdate 通知DNS更新
func (this *NSNodeDAO) NotifyDNSUpdate(tx *dbs.Tx, nodeId int64) error {
	// TODO 先什么都不做
	return nil
}
