//go:build !plus

package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
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

// CountAllLowerVersionNodesWithClusterId 计算单个集群中所有低于某个版本的节点数量
func (this *NSNodeDAO) CountAllLowerVersionNodesWithClusterId(tx *dbs.Tx, clusterId int64, os string, arch string, version string) (int64, error) {
	return this.Query(tx).
		State(NSNodeStateEnabled).
		Attr("clusterId", clusterId).
		Attr("isOn", true).
		Where("status IS NOT NULL").
		Where("JSON_EXTRACT(status, '$.os')=:os").
		Where("JSON_EXTRACT(status, '$.arch')=:arch").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("os", os).
		Param("arch", arch).
		Param("version", utils.VersionToLong(version)).
		Count()
}

// FindEnabledNodeIdWithUniqueId 根据唯一ID获取节点ID
func (this *NSNodeDAO) FindEnabledNodeIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", NSNodeStateEnabled).
		ResultPk().
		FindInt64Col(0)
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
		Attr("isOn", true).
		Where("clusterId IN (SELECT id FROM "+SharedNSClusterDAO.Table+" WHERE state=1)").
		Where("status IS NOT NULL").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("version", utils.VersionToLong(version)).
		Count()
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
