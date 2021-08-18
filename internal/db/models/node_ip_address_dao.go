package models

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"math"
	"strings"
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

// CreateAddress 创建IP地址
func (this *NodeIPAddressDAO) CreateAddress(tx *dbs.Tx, nodeId int64, role nodeconfigs.NodeRole, name string, ip string, canAccess bool, thresholdsJSON []byte) (addressId int64, err error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}

	op := NewNodeIPAddressOperator()
	op.NodeId = nodeId
	op.Role = role
	op.Name = name
	op.Ip = ip
	op.CanAccess = canAccess

	if len(thresholdsJSON) > 0 {
		op.Thresholds = thresholdsJSON
	} else {
		op.Thresholds = "[]"
	}

	op.State = NodeIPAddressStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	err = SharedNodeDAO.NotifyDNSUpdate(tx, nodeId)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// UpdateAddress 修改IP地址
func (this *NodeIPAddressDAO) UpdateAddress(tx *dbs.Tx, addressId int64, name string, ip string, canAccess bool, isOn bool, thresholdsJSON []byte) (err error) {
	if addressId <= 0 {
		return errors.New("invalid addressId")
	}

	op := NewNodeIPAddressOperator()
	op.Id = addressId
	op.Name = name
	op.Ip = ip
	op.CanAccess = canAccess
	op.IsOn = isOn

	if len(thresholdsJSON) > 0 {
		op.Thresholds = thresholdsJSON
	} else {
		op.Thresholds = "[]"
	}

	op.State = NodeIPAddressStateEnabled // 恢复状态
	err = this.Save(tx, op)
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
func (this *NodeIPAddressDAO) FindFirstNodeAccessIPAddress(tx *dbs.Tx, nodeId int64, role nodeconfigs.NodeRole) (string, error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	return this.Query(tx).
		Attr("nodeId", nodeId).
		Attr("role", role).
		State(NodeIPAddressStateEnabled).
		Attr("canAccess", true).
		Desc("order").
		AscPk().
		Result("ip").
		FindStringCol("")
}

// FindFirstNodeAccessIPAddressId 查找节点的第一个可访问的IP地址ID
func (this *NodeIPAddressDAO) FindFirstNodeAccessIPAddressId(tx *dbs.Tx, nodeId int64, role nodeconfigs.NodeRole) (int64, error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	return this.Query(tx).
		Attr("nodeId", nodeId).
		Attr("role", role).
		State(NodeIPAddressStateEnabled).
		Attr("canAccess", true).
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

// FireThresholds 触发阈值
func (this *NodeIPAddressDAO) FireThresholds(tx *dbs.Tx, role nodeconfigs.NodeRole, nodeId int64) error {

	ones, err := this.Query(tx).
		Attr("state", NodeIPAddressStateEnabled).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Attr("canAccess", true).
		Attr("isOn", true).
		FindAll()
	if err != nil {
		return err
	}
	for _, one := range ones {
		addr := one.(*NodeIPAddress)
		var thresholds = addr.DecodeThresholds()
		if len(thresholds) == 0 {
			continue
		}
		var isOk = true
		var summary = []string{}
		for _, threshold := range thresholds {
			if threshold.Value <= 0 || threshold.Duration <= 0 {
				continue
			}

			var value = float64(0)
			switch threshold.Item {
			case "avgRequests":
				value, err = SharedNodeValueDAO.SumValues(tx, role, nodeId, nodeconfigs.NodeValueItemRequests, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(threshold.Duration), threshold.DurationUnit)
				value = math.Round(value / 60)
				summary = append(summary, "平均请求数："+types.String(value)+"/s")
			case "avgTrafficOut":
				value, err = SharedNodeValueDAO.SumValues(tx, role, nodeId, nodeconfigs.NodeValueItemTrafficOut, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(threshold.Duration), threshold.DurationUnit)
				value = math.Round(value*100/1024/1024/60) / 100 // 100 = 两位小数
				summary = append(summary, "平均下行流量："+types.String(value)+"MB/s")
			case "avgTrafficIn":
				value, err = SharedNodeValueDAO.SumValues(tx, role, nodeId, nodeconfigs.NodeValueItemTrafficIn, "total", nodeconfigs.NodeValueSumMethodAvg, types.Int32(threshold.Duration), threshold.DurationUnit)
				value = math.Round(value*100/1024/1024/60) / 100 // 100 = 两位小数
				summary = append(summary, "平均上行流量："+types.String(value)+"MB/s")
			default:
				// TODO 支持更多
				err = errors.New("threshold item '" + threshold.Item + "' not supported")
			}
			if err != nil {
				return err
			}
			if !nodeconfigs.CompareNodeValue(threshold.Operator, value, float64(threshold.Value)) {
				isOk = false
			}
		}
		if isOk && addr.IsUp == 0 { // 新上线
			_, err := this.Query(tx).
				Pk(addr.Id).
				Set("isUp", true).
				Update()
			if err != nil {
				return err
			}

			clusterId, err := SharedNodeDAO.FindNodeClusterId(tx, nodeId)
			if err != nil {
				return err
			}
			err = SharedMessageDAO.CreateNodeMessage(tx, role, clusterId, nodeId, MessageTypeIPAddrUp, MessageLevelSuccess, "节点IP'"+addr.Ip+"'因为达到阈值而上线", "节点IP'"+addr.Ip+"'因为达到阈值而上线。"+strings.Join(summary, "，") + "。", maps.Map{
				"addrId": addr.Id,
			}.AsJSON())
			if err != nil {
				return err
			}

			err = this.NotifyUpdate(tx, int64(addr.Id))
			if err != nil {
				return err
			}
		} else if !isOk && addr.IsUp == 1 { // 新离线
			_, err := this.Query(tx).
				Pk(addr.Id).
				Set("isUp", false).
				Update()
			if err != nil {
				return err
			}

			clusterId, err := SharedNodeDAO.FindNodeClusterId(tx, nodeId)
			if err != nil {
				return err
			}
			err = SharedMessageDAO.CreateNodeMessage(tx, role, clusterId, nodeId, MessageTypeIPAddrDown, MessageLevelWarning, "节点IP'"+addr.Ip+"'因为达到阈值而下线", "节点IP'"+addr.Ip+"'因为达到阈值而下线。"+strings.Join(summary, "，") + "。", maps.Map{
				"addrId": addr.Id,
			}.AsJSON())
			if err != nil {
				return err
			}

			err = this.NotifyUpdate(tx, int64(addr.Id))
			if err != nil {
				return err
			}
		}
	}

	return nil
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
	}
	if err != nil {
		return err
	}
	return nil
}
