package models

import (
	"encoding/json"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/reporterconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
)

const (
	ReportNodeStateEnabled  = 1 // 已启用
	ReportNodeStateDisabled = 0 // 已禁用
)

type ReportNodeDAO dbs.DAO

func NewReportNodeDAO() *ReportNodeDAO {
	return dbs.NewDAO(&ReportNodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeReportNodes",
			Model:  new(ReportNode),
			PkName: "id",
		},
	}).(*ReportNodeDAO)
}

var SharedReportNodeDAO *ReportNodeDAO

func init() {
	dbs.OnReady(func() {
		SharedReportNodeDAO = NewReportNodeDAO()
	})
}

// EnableReportNode 启用条目
func (this *ReportNodeDAO) EnableReportNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ReportNodeStateEnabled).
		Update()
	return err
}

// DisableReportNode 禁用条目
func (this *ReportNodeDAO) DisableReportNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ReportNodeStateDisabled).
		Update()
	return err
}

// FindEnabledReportNode 查找启用中的条目
func (this *ReportNodeDAO) FindEnabledReportNode(tx *dbs.Tx, id int64) (*ReportNode, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ReportNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ReportNode), err
}

// FindReportNodeName 根据主键查找名称
func (this *ReportNodeDAO) FindReportNodeName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateReportNode 创建终端
func (this *ReportNodeDAO) CreateReportNode(tx *dbs.Tx, name string, location string, isp string, allowIPs []string, groupIds []int64) (int64, error) {
	uniqueId, err := this.GenUniqueId(tx)
	if err != nil {
		return 0, err
	}

	secret := rands.String(32)

	// 保存API Token
	err = SharedApiTokenDAO.CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleReport)
	if err != nil {
		return 0, err
	}

	op := NewReportNodeOperator()
	op.UniqueId = uniqueId
	op.Secret = secret
	op.Name = name
	op.Location = location
	op.Isp = isp

	if len(allowIPs) > 0 {
		allowIPSJSON, err := json.Marshal(allowIPs)
		if err != nil {
			return 0, err
		}
		op.AllowIPs = allowIPSJSON
	} else {
		op.AllowIPs = "[]"
	}

	if len(groupIds) > 0 {
		groupIdsJSON, err := json.Marshal(groupIds)
		if err != nil {
			return 0, err
		}
		op.GroupIds = groupIdsJSON
	} else {
		op.GroupIds = "[]"
	}

	op.IsOn = true
	op.State = ReportNodeStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateReportNode 修改终端
func (this *ReportNodeDAO) UpdateReportNode(tx *dbs.Tx, nodeId int64, name string, location string, isp string, allowIPs []string, groupIds []int64, isOn bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	op := NewReportNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.Location = location
	op.Isp = isp

	if len(allowIPs) > 0 {
		allowIPSJSON, err := json.Marshal(allowIPs)
		if err != nil {
			return err
		}
		op.AllowIPs = allowIPSJSON
	} else {
		op.AllowIPs = "[]"
	}

	if len(groupIds) > 0 {
		groupIdsJSON, err := json.Marshal(groupIds)
		if err != nil {
			return err
		}
		op.GroupIds = groupIdsJSON
	} else {
		op.GroupIds = "[]"
	}

	op.IsOn = isOn
	return this.Save(tx, op)
}

// CountAllEnabledReportNodes 计算终端数量
func (this *ReportNodeDAO) CountAllEnabledReportNodes(tx *dbs.Tx, groupId int64, keyword string) (int64, error) {
	var query = this.Query(tx).
		State(ReportNodeStateEnabled)
	if groupId > 0 {
		query.JSONContains("groupIds", types.String(groupId))
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR location LIKE :keyword OR isp LIKE :keyword OR allowIPs LIKE :keyword OR (status IS NOT NULL AND JSON_EXTRACT(status, 'ip') LIKE :keyword))")
		query.Param("keyword", dbutils.QuoteLike(keyword))
	}
	return query.Count()
}

// CountAllEnabledAndOnReportNodes 计算可用的终端数量
func (this *ReportNodeDAO) CountAllEnabledAndOnReportNodes(tx *dbs.Tx) (int64, error) {
	var query = this.Query(tx).
		Attr("isOn", true).
		State(ReportNodeStateEnabled)
	return query.Count()
}

// ListEnabledReportNodes 列出单页终端
func (this *ReportNodeDAO) ListEnabledReportNodes(tx *dbs.Tx, groupId int64, keyword string, offset int64, size int64) (result []*ReportNode, err error) {
	var query = this.Query(tx).
		State(ReportNodeStateEnabled)
	if groupId > 0 {
		query.JSONContains("groupIds", types.String(groupId))
	}
	if len(keyword) > 0 {
		query.Where(`(
	name LIKE :keyword 
	OR location LIKE :keyword 
	OR isp LIKE :keyword 
	OR allowIPs LIKE :keyword 
	OR (status IS NOT NULL 
		AND (
			JSON_EXTRACT(status, '$.ip') LIKE :keyword) 
			OR (LENGTH(location)=0 AND JSON_EXTRACT(status, '$.location') LIKE :keyword) 
			OR (LENGTH(isp)=0 AND JSON_EXTRACT(status, '$.isp') LIKE :keyword)
       ))`)
		query.Param("keyword", dbutils.QuoteLike(keyword))
	}
	query.Slice(&result)
	_, err = query.Asc("isActive").
		Offset(offset).
		Limit(size).
		DescPk().
		FindAll()
	return
}

// GenUniqueId 生成唯一ID
func (this *ReportNodeDAO) GenUniqueId(tx *dbs.Tx) (string, error) {
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

// UpdateNodeActive 修改节点活跃状态
func (this *ReportNodeDAO) UpdateNodeActive(tx *dbs.Tx, nodeId int64, isActive bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("isActive", isActive).
		Update()
	return err
}

// FindNodeActive 检查节点活跃状态
func (this *ReportNodeDAO) FindNodeActive(tx *dbs.Tx, nodeId int64) (bool, error) {
	isActive, err := this.Query(tx).
		Pk(nodeId).
		Result("isActive").
		FindIntCol(0)
	if err != nil {
		return false, err
	}
	return isActive == 1, nil
}

// FindEnabledNodeIdWithUniqueId 根据唯一ID获取节点ID
func (this *ReportNodeDAO) FindEnabledNodeIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", ReportNodeStateEnabled).
		ResultPk().
		FindInt64Col(0)
}

// UpdateNodeStatus 更改节点状态
func (this ReportNodeDAO) UpdateNodeStatus(tx *dbs.Tx, nodeId int64, nodeStatus *reporterconfigs.Status) error {
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

// ComposeConfig 组合配置
func (this *ReportNodeDAO) ComposeConfig(tx *dbs.Tx, nodeId int64) (*reporterconfigs.NodeConfig, error) {
	node, err := this.FindEnabledReportNode(tx, nodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}
	var config = &reporterconfigs.NodeConfig{
		Id: int64(node.Id),
	}
	return config, nil
}

// FindNodeAllowIPs 查询节点允许的IP
func (this *ReportNodeDAO) FindNodeAllowIPs(tx *dbs.Tx, nodeId int64) ([]string, error) {
	node, err := this.Query(tx).
		Pk(nodeId).
		Result("allowIPs").
		Find()
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}
	return node.(*ReportNode).DecodeAllowIPs(), nil
}

// CountAllLowerVersionNodes 计算所有节点中低于某个版本的节点数量
func (this *ReportNodeDAO) CountAllLowerVersionNodes(tx *dbs.Tx, version string) (int64, error) {
	return this.Query(tx).
		State(ReportNodeStateEnabled).
		Where("status IS NOT NULL").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("version", utils.VersionToLong(version)).
		Count()
}
