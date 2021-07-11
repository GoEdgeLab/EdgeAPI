package nameservers

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
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
func (this *NSNodeDAO) DisableNSNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSNodeStateDisabled).
		Update()

	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, id)
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
		query.Where("JSON_EXTRACT(status, '$.isActive') AND UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')<=60")
	case configutils.BoolStateNo:
		query.Where("(status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>60)")
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
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
		query.Where("JSON_EXTRACT(status, '$.isActive') AND UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')<=60")
	case configutils.BoolStateNo:
		query.Where("(status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>60)")
	}

	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
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
	err = models.SharedApiTokenDAO.CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleDNS)
	if err != nil {
		return
	}

	op := NewNSNodeOperator()
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
	op := NewNSNodeOperator()
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
func (this *NSNodeDAO) FindNodeInstallStatus(tx *dbs.Tx, nodeId int64) (*models.NodeInstallStatus, error) {
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
	isInstalled := node.(*NSNode).IsInstalled == 1
	if len(installStatus) == 0 {
		return models.NewNodeInstallStatus(), nil
	}

	status := &models.NodeInstallStatus{}
	err = json.Unmarshal([]byte(installStatus), status)
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
func (this NSNodeDAO) UpdateNodeStatus(tx *dbs.Tx, nodeId int64, statusJSON []byte) error {
	if statusJSON == nil {
		return nil
	}
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("status", string(statusJSON)).
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

	config := &dnsconfigs.NSNodeConfig{
		Id:        int64(node.Id),
		ClusterId: int64(node.ClusterId),
	}

	if len(cluster.AccessLog) > 0 {
		ref := &dnsconfigs.AccessLogRef{}
		err = json.Unmarshal([]byte(cluster.AccessLog), ref)
		if err != nil {
			return nil, err
		}
		config.AccessLogRef = ref
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
