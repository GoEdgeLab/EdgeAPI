package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"math"
	"time"
)

const (
	IPItemStateEnabled  = 1 // 已启用
	IPItemStateDisabled = 0 // 已禁用
)

type IPItemType = string

const (
	IPItemTypeIPv4 IPItemType = "ipv4" // IPv4
	IPItemTypeIPv6 IPItemType = "ipv6" // IPv6
	IPItemTypeAll  IPItemType = "all"  // 所有IP
)

type IPItemDAO dbs.DAO

func NewIPItemDAO() *IPItemDAO {
	return dbs.NewDAO(&IPItemDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeIPItems",
			Model:  new(IPItem),
			PkName: "id",
		},
	}).(*IPItemDAO)
}

var SharedIPItemDAO *IPItemDAO

func init() {
	dbs.OnReady(func() {
		SharedIPItemDAO = NewIPItemDAO()
	})
}

// EnableIPItem 启用条目
func (this *IPItemDAO) EnableIPItem(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPItemStateEnabled).
		Update()
	return err
}

// DisableIPItem 禁用条目
func (this *IPItemDAO) DisableIPItem(tx *dbs.Tx, id int64) error {
	version, err := SharedIPListDAO.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(id).
		Set("state", IPItemStateDisabled).
		Set("version", version).
		Update()

	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, id)
}

// FindEnabledIPItem 查找启用中的条目
func (this *IPItemDAO) FindEnabledIPItem(tx *dbs.Tx, id int64) (*IPItem, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", IPItemStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*IPItem), err
}

// CreateIPItem 创建IP
func (this *IPItemDAO) CreateIPItem(tx *dbs.Tx, listId int64, ipFrom string, ipTo string, expiredAt int64, reason string, itemType IPItemType, eventLevel string) (int64, error) {
	version, err := SharedIPListDAO.IncreaseVersion(tx)
	if err != nil {
		return 0, err
	}

	op := NewIPItemOperator()
	op.ListId = listId
	op.IpFrom = ipFrom
	op.IpTo = ipTo
	op.IpFromLong = utils.IP2Long(ipFrom)
	op.IpToLong = utils.IP2Long(ipTo)
	op.Reason = reason
	op.Type = itemType
	op.EventLevel = eventLevel
	op.Version = version
	if expiredAt < 0 {
		expiredAt = 0
	}
	op.ExpiredAt = expiredAt
	op.State = IPItemStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	itemId := types.Int64(op.Id)

	err = this.NotifyUpdate(tx, itemId)
	if err != nil {
		return 0, err
	}
	return itemId, nil
}

// UpdateIPItem 修改IP
func (this *IPItemDAO) UpdateIPItem(tx *dbs.Tx, itemId int64, ipFrom string, ipTo string, expiredAt int64, reason string, itemType IPItemType, eventLevel string) error {
	if itemId <= 0 {
		return errors.New("invalid itemId")
	}

	listId, err := this.Query(tx).
		Pk(itemId).
		Result("listId").
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if listId == 0 {
		return errors.New("not found")
	}

	version, err := SharedIPListDAO.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	op := NewIPItemOperator()
	op.Id = itemId
	op.IpFrom = ipFrom
	op.IpTo = ipTo
	op.IpFromLong = utils.IP2Long(ipFrom)
	op.IpToLong = utils.IP2Long(ipTo)
	op.Reason = reason
	op.Type = itemType
	op.EventLevel = eventLevel
	if expiredAt < 0 {
		expiredAt = 0
	}
	op.ExpiredAt = expiredAt
	op.Version = version
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, itemId)
}

// CountIPItemsWithListId 计算IP数量
func (this *IPItemDAO) CountIPItemsWithListId(tx *dbs.Tx, listId int64) (int64, error) {
	return this.Query(tx).
		State(IPItemStateEnabled).
		Attr("listId", listId).
		Count()
}

// ListIPItemsWithListId 查找IP列表
func (this *IPItemDAO) ListIPItemsWithListId(tx *dbs.Tx, listId int64, offset int64, size int64) (result []*IPItem, err error) {
	_, err = this.Query(tx).
		State(IPItemStateEnabled).
		Attr("listId", listId).
		DescPk().
		Slice(&result).
		Offset(offset).
		Limit(size).
		FindAll()
	return
}

// ListIPItemsAfterVersion 根据版本号查找IP列表
func (this *IPItemDAO) ListIPItemsAfterVersion(tx *dbs.Tx, version int64, size int64) (result []*IPItem, err error) {
	_, err = this.Query(tx).
		// 这里不要设置状态参数，因为我们要知道哪些是删除的
		Gt("version", version).
		Where("(expiredAt=0 OR expiredAt>:expiredAt)").
		Param("expiredAt", time.Now().Unix()).
		Asc("version").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindItemListId 查找IPItem对应的列表ID
func (this *IPItemDAO) FindItemListId(tx *dbs.Tx, itemId int64) (int64, error) {
	return this.Query(tx).
		Pk(itemId).
		Result("listId").
		FindInt64Col(0)
}

// FindEnabledItemContainsIP 查找包含某个IP的Item
func (this *IPItemDAO) FindEnabledItemContainsIP(tx *dbs.Tx, listId int64, ip uint64) (*IPItem, error) {
	query := this.Query(tx).
		Attr("listId", listId).
		State(IPItemStateEnabled)
	if ip > math.MaxUint32 {
		query.Where("(type='all' OR ipFromLong=:ip)")
	} else {
		query.Where("(type='all' OR ipFromLong=:ip OR (ipToLong>0 AND ipFromLong<=:ip AND ipToLong>=:ip))").
			Param("ip", ip)
	}
	one, err := query.Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*IPItem), nil
}

// ExistsEnabledItem 检查IP是否存在
func (this *IPItemDAO) ExistsEnabledItem(tx *dbs.Tx, itemId int64) (bool, error) {
	return this.Query(tx).
		Pk(itemId).
		State(IPItemStateEnabled).
		Exist()
}

// NotifyUpdate 通知更新
func (this *IPItemDAO) NotifyUpdate(tx *dbs.Tx, itemId int64) error {
	// 获取ListId
	listId, err := this.FindItemListId(tx, itemId)
	if err != nil {
		return err
	}

	if listId == 0 {
		return nil
	}

	httpFirewallPolicyIds, err := SharedHTTPFirewallPolicyDAO.FindEnabledFirewallPolicyIdsWithIPListId(tx, listId)
	if err != nil {
		return err
	}
	resultClusterIds := []int64{}
	for _, policyId := range httpFirewallPolicyIds {
		// 集群
		clusterIds, err := SharedNodeClusterDAO.FindAllEnabledNodeClusterIdsWithHTTPFirewallPolicyId(tx, policyId)
		if err != nil {
			return err
		}
		for _, clusterId := range clusterIds {
			if !lists.ContainsInt64(resultClusterIds, clusterId) {
				resultClusterIds = append(resultClusterIds, clusterId)
			}
		}

		// 服务
		webIds, err := SharedHTTPWebDAO.FindAllWebIdsWithHTTPFirewallPolicyId(tx, policyId)
		if err != nil {
			return err
		}
		if len(webIds) > 0 {
			for _, webId := range webIds {
				serverId, err := SharedServerDAO.FindEnabledServerIdWithWebId(tx, webId)
				if err != nil {
					return err
				}
				if serverId > 0 {
					clusterId, err := SharedServerDAO.FindServerClusterId(tx, serverId)
					if err != nil {
						return err
					}
					if !lists.ContainsInt64(resultClusterIds, clusterId) {
						resultClusterIds = append(resultClusterIds, clusterId)
					}
				}
			}
		}
	}

	if len(resultClusterIds) > 0 {
		for _, clusterId := range resultClusterIds {
			err = SharedNodeTaskDAO.CreateClusterTask(tx, clusterId, NodeTaskTypeIPItemChanged)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
