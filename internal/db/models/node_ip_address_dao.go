package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	NodeIPAddressStateEnabled  = 1 // 已启用
	NodeIPAddressStateDisabled = 0 // 已禁用
)

type NodeIPAddressDAO dbs.DAO

func NewNodeIPAddressDAO() *NodeIPAddressDAO {
	return dbs.NewDAO(&NodeIPAddressDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeIPAddresses",
			Model:  new(NodeIPAddress),
			PkName: "id",
		},
	}).(*NodeIPAddressDAO)
}

var SharedNodeIPAddressDAO *NodeIPAddressDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeIPAddressDAO = NewNodeIPAddressDAO()
	})
}

// EnableAddress 启用条目
func (this *NodeIPAddressDAO) EnableAddress(tx *dbs.Tx, addressId int64) (err error) {
	_, err = this.Query(tx).
		Pk(addressId).
		Set("state", NodeIPAddressStateEnabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, addressId)
}

// DisableAddress 禁用IP地址
func (this *NodeIPAddressDAO) DisableAddress(tx *dbs.Tx, addressId int64) (err error) {
	_, err = this.Query(tx).
		Pk(addressId).
		Set("state", NodeIPAddressStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, addressId)
}

// DisableAllAddressesWithNodeId 禁用节点的所有的IP地址
func (this *NodeIPAddressDAO) DisableAllAddressesWithNodeId(tx *dbs.Tx, nodeId int64, role nodeconfigs.NodeRole) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	_, err := this.Query(tx).
		Attr("nodeId", nodeId).
		Attr("role", role).
		Set("state", NodeIPAddressStateDisabled).
		Update()
	if err != nil {
		return err
	}

	return SharedNodeDAO.NotifyDNSUpdate(tx, nodeId)
}

// FindEnabledAddress 查找启用中的IP地址
func (this *NodeIPAddressDAO) FindEnabledAddress(tx *dbs.Tx, id int64) (*NodeIPAddress, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeIPAddressStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeIPAddress), err
}

// FindAddressName 根据主键查找名称
func (this *NodeIPAddressDAO) FindAddressName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindAddressIsHealthy 判断IP地址是否健康
func (this *NodeIPAddressDAO) FindAddressIsHealthy(tx *dbs.Tx, addressId int64) (isHealthy bool, err error) {
	if addressId <= 0 {
		return false, nil
	}
	one, err := this.Query(tx).
		Pk(addressId).
		Result("isHealthy").
		Find()
	if err != nil || one == nil {
		return false, err
	}
	var addr = one.(*NodeIPAddress)
	return addr.IsHealthy, nil
}

// CreateAddress 创建IP地址
func (this *NodeIPAddressDAO) CreateAddress(tx *dbs.Tx, adminId int64, nodeId int64, role nodeconfigs.NodeRole, name string, ip string, canAccess bool, isUp bool, groupId int64) (addressId int64, err error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}

	var op = NewNodeIPAddressOperator()
	op.NodeId = nodeId
	op.Role = role
	op.Name = name
	op.Ip = ip
	op.CanAccess = canAccess
	op.IsUp = isUp
	op.GroupId = groupId

	op.State = NodeIPAddressStateEnabled
	addressId, err = this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = SharedNodeDAO.NotifyDNSUpdate(tx, nodeId)
	if err != nil {
		return 0, err
	}

	// 创建日志
	err = SharedNodeIPAddressLogDAO.CreateLog(tx, adminId, addressId, "创建IP")
	if err != nil {
		return 0, err
	}

	return addressId, nil
}

// UpdateAddress 修改IP地址
func (this *NodeIPAddressDAO) UpdateAddress(tx *dbs.Tx, adminId int64, addressId int64, name string, ip string, canAccess bool, isOn bool, isUp bool) (err error) {
	if addressId <= 0 {
		return errors.New("invalid addressId")
	}

	op := NewNodeIPAddressOperator()
	op.Id = addressId
	op.Name = name
	op.Ip = ip
	op.CanAccess = canAccess
	op.IsOn = isOn
	op.IsUp = isUp

	op.State = NodeIPAddressStateEnabled // 恢复状态
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	// 创建日志
	err = SharedNodeIPAddressLogDAO.CreateLog(tx, adminId, addressId, "修改IP")
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, addressId)
}

// UpdateAddressIP 修改IP地址中的IP
func (this *NodeIPAddressDAO) UpdateAddressIP(tx *dbs.Tx, addressId int64, ip string) error {
	if addressId <= 0 {
		return errors.New("invalid addressId")
	}
	op := NewNodeIPAddressOperator()
	op.Id = addressId
	op.Ip = ip
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, addressId)
}

// UpdateAddressNodeId 修改IP地址所属节点
func (this *NodeIPAddressDAO) UpdateAddressNodeId(tx *dbs.Tx, addressId int64, nodeId int64) error {
	_, err := this.Query(tx).
		Pk(addressId).
		Set("nodeId", nodeId).
		Set("state", NodeIPAddressStateEnabled). // 恢复状态
		Update()
	if err != nil {
		return err
	}

	err = SharedNodeDAO.NotifyDNSUpdate(tx, nodeId)
	if err != nil {
		return err
	}
	return nil
}

// FindAllEnabledAddressesWithNode 查找节点的所有的IP地址
func (this *NodeIPAddressDAO) FindAllEnabledAddressesWithNode(tx *dbs.Tx, nodeId int64, role nodeconfigs.NodeRole) (result []*NodeIPAddress, err error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	_, err = this.Query(tx).
		Attr("nodeId", nodeId).
		Attr("role", role).
		State(NodeIPAddressStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindFirstNodeAccessIPAddress 查找节点的第一个可访问的IP地址
func (this *NodeIPAddressDAO) FindFirstNodeAccessIPAddress(tx *dbs.Tx, nodeId int64, mustUp bool, role nodeconfigs.NodeRole) (ip string, addrId int64, err error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	var query = this.Query(tx).
		Attr("nodeId", nodeId).
		Attr("role", role).
		State(NodeIPAddressStateEnabled).
		Attr("canAccess", true).
		Attr("isOn", true)
	if mustUp {
		query.Attr("isUp", true)
	}
	one, err := query.
		Desc("order").
		AscPk().
		Result("id", "ip").
		Find()
	if err != nil {
		return "", 0, err
	}
	if one == nil {
		return
	}

	var addr = one.(*NodeIPAddress)
	return addr.Ip, int64(addr.Id), nil
}

// FindFirstNodeAccessIPAddressId 查找节点的第一个可访问的IP地址ID
func (this *NodeIPAddressDAO) FindFirstNodeAccessIPAddressId(tx *dbs.Tx, nodeId int64, mustUp bool, role nodeconfigs.NodeRole) (int64, error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	var query = this.Query(tx).
		Attr("nodeId", nodeId).
		Attr("role", role).
		State(NodeIPAddressStateEnabled).
		Attr("canAccess", true).
		Attr("isOn", true)
	if mustUp {
		query.Attr("isUp", true)
	}
	return query.
		Desc("order").
		AscPk().
		Result("id").
		FindInt64Col(0)
}

// FindNodeAccessAndUpIPAddresses 查找节点所有的可访问的IP地址
func (this *NodeIPAddressDAO) FindNodeAccessAndUpIPAddresses(tx *dbs.Tx, nodeId int64, role nodeconfigs.NodeRole) (result []*NodeIPAddress, err error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	_, err = this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		State(NodeIPAddressStateEnabled).
		Attr("canAccess", true).
		Attr("isOn", true).
		Attr("isUp", true).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledIPAddresses 计算IP地址数量
// TODO 目前支持边缘节点，将来支持NS节点
func (this *NodeIPAddressDAO) CountAllEnabledIPAddresses(tx *dbs.Tx, role string, nodeClusterId int64, upState configutils.BoolState, keyword string) (int64, error) {
	var query = this.Query(tx).
		State(NodeIPAddressStateEnabled).
		Attr("role", role)

	// 集群
	if nodeClusterId > 0 {
		query.Where("nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE (clusterId=:clusterId OR JSON_CONTAINS(secondaryClusterIds, :clusterIdString)) AND state=1)").
			Param("clusterId", nodeClusterId).
			Param("clusterIdString", types.String(nodeClusterId))
	} else {
		query.Where("nodeId IN (SELECT id FROM " + SharedNodeDAO.Table + " WHERE state=1 AND clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1))")
	}

	// 在线状态
	switch upState {
	case configutils.BoolStateYes:
		query.Attr("isUp", 1)
	case configutils.BoolStateNo:
		query.Attr("isUp", 0)
	}

	// 关键词
	if len(keyword) > 0 {
		query.Where("(ip LIKE :keyword OR name LIKE :keyword OR description LIKE :keyword OR nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE state=1 AND name LIKE :keyword))").
			Param("keyword", dbutils.QuoteLike(keyword))
	}

	return query.Count()
}

// ListEnabledIPAddresses 列出单页的IP地址
func (this *NodeIPAddressDAO) ListEnabledIPAddresses(tx *dbs.Tx, role string, nodeClusterId int64, upState configutils.BoolState, keyword string, offset int64, size int64) (result []*NodeIPAddress, err error) {
	var query = this.Query(tx).
		State(NodeIPAddressStateEnabled).
		Attr("role", role)

	// 集群
	if nodeClusterId > 0 {
		query.Where("nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE (clusterId=:clusterId OR JSON_CONTAINS(secondaryClusterIds, :clusterIdString)) AND state=1)").
			Param("clusterId", nodeClusterId).
			Param("clusterIdString", types.String(nodeClusterId))
	} else {
		query.Where("nodeId IN (SELECT id FROM " + SharedNodeDAO.Table + " WHERE state=1 AND clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1))")
	}

	// 在线状态
	switch upState {
	case configutils.BoolStateYes:
		query.Attr("isUp", 1)
	case configutils.BoolStateNo:
		query.Attr("isUp", 0)
	}

	// 关键词
	if len(keyword) > 0 {
		query.Where("(ip LIKE :keyword OR name LIKE :keyword OR description LIKE :keyword OR nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE state=1 AND name LIKE :keyword))").
			Param("keyword", dbutils.QuoteLike(keyword))
	}

	_, err = query.Offset(offset).
		Limit(size).
		Asc("isUp").
		Desc("nodeId").
		Slice(&result).
		FindAll()
	return
}

// FindAllAccessibleIPAddressesWithClusterId 列出所有的正在启用的IP地址
func (this *NodeIPAddressDAO) FindAllAccessibleIPAddressesWithClusterId(tx *dbs.Tx, role string, clusterId int64) (result []*NodeIPAddress, err error) {
	_, err = this.Query(tx).
		State(NodeIPAddressStateEnabled).
		Attr("role", role).
		Attr("isOn", true).
		Attr("canAccess", true).
		Where("nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE state=1 AND clusterId=:clusterId)").
		Param("clusterId", clusterId).
		Slice(&result).
		FindAll()
	return
}

// CountAllAccessibleIPAddressesWithClusterId 计算集群中的可用IP地址数量
func (this *NodeIPAddressDAO) CountAllAccessibleIPAddressesWithClusterId(tx *dbs.Tx, role string, clusterId int64) (count int64, err error) {
	return this.Query(tx).
		State(NodeIPAddressStateEnabled).
		Attr("role", role).
		Attr("isOn", true).
		Attr("canAccess", true).
		Where("nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE state=1 AND clusterId=:clusterId)").
		Param("clusterId", clusterId).
		Count()
}

// ListAccessibleIPAddressesWithClusterId 列出单页集群中的可用IP地址
func (this *NodeIPAddressDAO) ListAccessibleIPAddressesWithClusterId(tx *dbs.Tx, role string, clusterId int64, offset int64, size int64) (result []*NodeIPAddress, err error) {
	_, err = this.Query(tx).
		State(NodeIPAddressStateEnabled).
		Attr("role", role).
		Attr("isOn", true).
		Attr("canAccess", true).
		Where("nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE state=1 AND clusterId=:clusterId)").
		Param("clusterId", clusterId).
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// UpdateAddressConnectivity 设置连通性数据
func (this *NodeIPAddressDAO) UpdateAddressConnectivity(tx *dbs.Tx, addrId int64, connectivity *nodeconfigs.Connectivity) error {
	connectivityJSON, err := json.Marshal(connectivity)
	if err != nil {
		return err
	}
	return this.Query(tx).
		Pk(addrId).
		Set("connectivity", connectivityJSON).
		UpdateQuickly()
}

// UpdateAddressIsUp 设置IP地址在线状态
func (this *NodeIPAddressDAO) UpdateAddressIsUp(tx *dbs.Tx, addressId int64, isUp bool) error {
	var err = this.Query(tx).
		Pk(addressId).
		Set("isUp", isUp).
		Set("countUp", 0).
		Set("countDown", 0).
		Set("backupIP", "").
		Set("backupThresholdId", 0).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, addressId)
}

// UpdateAddressBackupIP 设置备用IP
func (this *NodeIPAddressDAO) UpdateAddressBackupIP(tx *dbs.Tx, addressId int64, thresholdId int64, ip string) error {
	if addressId <= 0 {
		return errors.New("invalid addressId")
	}
	var op = NewNodeIPAddressOperator()
	op.IsUp = true // IP必须在线备用IP才会有用
	op.Id = addressId
	op.BackupThresholdId = thresholdId
	op.BackupIP = ip
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, addressId)
}

// UpdateAddressHealthCount 计算IP健康状态
func (this *NodeIPAddressDAO) UpdateAddressHealthCount(tx *dbs.Tx, addrId int64, newIsUp bool, maxUp int, maxDown int, autoUpDown bool) (changed bool, err error) {
	if addrId <= 0 {
		return false, errors.New("invalid address id")
	}
	one, err := this.Query(tx).
		Pk(addrId).
		Result("isHealthy", "isUp", "countUp", "countDown").
		Find()
	if err != nil {
		return false, err
	}
	if one == nil {
		return false, nil
	}
	var oldIsHealthy = one.(*NodeIPAddress).IsHealthy
	var oldIsUp = one.(*NodeIPAddress).IsUp

	// 如果新老状态一致，则不做任何事情
	if oldIsHealthy == newIsUp {
		// 如果自动上下线，则健康状况和是否在线保持一致
		if autoUpDown {
			if oldIsUp != oldIsHealthy {
				err = this.Query(tx).
					Pk(addrId).
					Set("isUp", newIsUp).
					UpdateQuickly()
				if err != nil {
					return false, err
				}
				err = this.NotifyUpdate(tx, addrId)
				if err != nil {
					return false, err
				}

				// 创建日志
				if newIsUp {
					err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, addrId, "健康检查上线")
				} else {
					err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, addrId, "健康检查下线")
				}
				if err != nil {
					return true, err
				}

				return true, nil
			}
		}
		return false, nil
	}

	var countUp = int(one.(*NodeIPAddress).CountUp)
	var countDown = int(one.(*NodeIPAddress).CountDown)

	var op = NewNodeIPAddressOperator()
	op.Id = addrId

	if newIsUp {
		countUp++
		countDown = 0

		if countUp >= maxUp {
			changed = true
			if autoUpDown {
				op.IsUp = true
			}
			op.IsHealthy = true
		}
	} else {
		countDown++
		countUp = 0

		if countDown >= maxDown {
			changed = true
			if autoUpDown {
				op.IsUp = false
			}
			op.IsHealthy = false
		}
	}

	op.CountUp = countUp
	op.CountDown = countDown
	err = this.Save(tx, op)
	if err != nil {
		return false, err
	}

	if changed {
		err = this.NotifyUpdate(tx, addrId)
		if err != nil {
			return true, err
		}

		// 创建日志
		if autoUpDown {
			if newIsUp {
				err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, addrId, "健康检查上线")
			} else {
				err = SharedNodeIPAddressLogDAO.CreateLog(tx, 0, addrId, "健康检查下线")
			}
		}
	}

	return
}

// NotifyUpdate 通知更新
func (this *NodeIPAddressDAO) NotifyUpdate(tx *dbs.Tx, addressId int64) error {
	address, err := this.Query(tx).
		Pk(addressId).
		Result("nodeId", "role").
		Find()
	if err != nil {
		return err
	}
	if address == nil {
		return nil
	}
	var nodeId = int64(address.(*NodeIPAddress).NodeId)
	if nodeId == 0 {
		return nil
	}
	var role = address.(*NodeIPAddress).Role
	switch role {
	case nodeconfigs.NodeRoleNode:
		err = dns.SharedDNSTaskDAO.CreateNodeTask(tx, nodeId, dns.DNSTaskTypeNodeChange)
		if err != nil {
			return err
		}

		// 检查是否为L2以上级别
		level, err := SharedNodeDAO.FindNodeLevel(tx, nodeId)
		if err != nil {
			return err
		}
		if level > 1 {
			err = SharedNodeDAO.NotifyLevelUpdate(tx, nodeId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
