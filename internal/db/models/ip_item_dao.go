package models

import (
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/iputils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"net"
	"strings"
	"time"
)

const (
	IPItemStateEnabled  = 1 // 已启用
	IPItemStateDisabled = 0 // 已禁用
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			var ticker = time.NewTicker(1 * time.Minute)
			for range ticker.C {
				err := SharedIPItemDAO.CleanExpiredIPItems(nil)
				if err != nil {
					remotelogs.Error("IPItemDAO", "clean expired ip items failed: "+err.Error())
				}
			}
		})
	})
}

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
func (this *IPItemDAO) DisableIPItem(tx *dbs.Tx, itemId int64, sourceUserId int64) error {
	version, err := SharedIPListDAO.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	var query = this.Query(tx)

	// 检查权限
	if sourceUserId > 0 {
		query.Where("(sourceUserId=:sourceUserId OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + "  WHERE userId=:sourceUserId AND state=1))")
		query.Param("sourceUserId", sourceUserId)
	}

	_, err = query.
		Pk(itemId).
		Set("state", IPItemStateDisabled).
		Set("version", version).
		Update()

	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, itemId)
}

// DisableIPItemsWithIP 禁用某个IP相关条目
func (this *IPItemDAO) DisableIPItemsWithIP(tx *dbs.Tx, ipFrom string, ipTo string, sourceUserId int64, listId int64) error {
	if len(ipFrom) == 0 {
		return errors.New("invalid 'ipFrom'")
	}

	var query = this.Query(tx).
		Result("id", "listId").
		Attr("ipFrom", ipFrom).
		Attr("ipTo", ipTo).
		State(IPItemStateEnabled)

	if listId > 0 {
		query.Attr("listId", listId)
	}

	if sourceUserId > 0 {
		query.Attr("sourceUserId", sourceUserId)
	}

	ones, err := query.FindAll()
	if err != nil {
		return err
	}

	var itemIds = []int64{}
	for _, one := range ones {
		var item = one.(*IPItem)
		var itemId = int64(item.Id)
		itemIds = append(itemIds, itemId)
	}

	for _, itemId := range itemIds {
		version, err := SharedIPListDAO.IncreaseVersion(tx)
		if err != nil {
			return err
		}

		_, err = this.Query(tx).
			Pk(itemId).
			Set("state", IPItemStateDisabled).
			Set("version", version).
			Update()
		if err != nil {
			return err
		}
	}

	if len(itemIds) > 0 {
		return this.NotifyUpdate(tx, itemIds[len(itemIds)-1])
	}
	return nil
}

// DisableIPItemsWithIPValue 禁用某个IP相关条目
func (this *IPItemDAO) DisableIPItemsWithIPValue(tx *dbs.Tx, value string, sourceUserId int64, listId int64) error {
	if len(value) == 0 {
		return errors.New("invalid 'value'")
	}

	var query = this.Query(tx).
		Result("id", "listId").
		Attr("value", value).
		State(IPItemStateEnabled)

	if listId > 0 {
		query.Attr("listId", listId)
	}

	if sourceUserId > 0 {
		query.Attr("sourceUserId", sourceUserId)
	}

	ones, err := query.FindAll()
	if err != nil {
		return err
	}

	var itemIds = []int64{}
	for _, one := range ones {
		var item = one.(*IPItem)
		var itemId = int64(item.Id)
		itemIds = append(itemIds, itemId)
	}

	for _, itemId := range itemIds {
		version, err := SharedIPListDAO.IncreaseVersion(tx)
		if err != nil {
			return err
		}

		_, err = this.Query(tx).
			Pk(itemId).
			Set("state", IPItemStateDisabled).
			Set("version", version).
			Update()
		if err != nil {
			return err
		}
	}

	if len(itemIds) > 0 {
		return this.NotifyUpdate(tx, itemIds[len(itemIds)-1])
	}
	return nil
}

// DisableIPItemsWithListId 禁用某个IP名单内的所有IP
func (this *IPItemDAO) DisableIPItemsWithListId(tx *dbs.Tx, listId int64) error {
	for {
		ones, err := this.Query(tx).
			ResultPk().
			Attr("listId", listId).
			State(IPItemStateEnabled).
			Limit(1000).
			FindAll()
		if err != nil {
			return err
		}
		if len(ones) == 0 {
			break
		}
		for _, one := range ones {
			var itemId = one.(*IPItem).Id
			version, err := SharedIPListDAO.IncreaseVersion(tx)
			if err != nil {
				return err
			}
			err = this.Query(tx).
				Pk(itemId).
				State(IPItemStateEnabled).
				Set("version", version).
				Set("state", IPItemStateDisabled).
				UpdateQuickly()
			if err != nil {
				return err
			}
		}
	}
	return nil
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

// DeleteOldItem 根据IP删除以前的旧记录
func (this *IPItemDAO) DeleteOldItem(tx *dbs.Tx, listId int64, ipFrom string, ipTo string) error {
	ones, err := this.Query(tx).
		ResultPk().
		UseIndex("ipFrom").
		Attr("listId", listId).
		Attr("ipFrom", ipFrom).
		Attr("ipTo", ipTo).
		Attr("state", IPItemStateEnabled).
		FindAll()
	if err != nil {
		return err
	}

	for _, one := range ones {
		var itemId = int64(one.(*IPItem).Id)
		version, err := SharedIPListDAO.IncreaseVersion(tx)
		if err != nil {
			return err
		}

		err = this.Query(tx).
			Pk(itemId).
			Set("version", version).
			Set("state", IPItemStateDisabled).
			UpdateQuickly()
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteOldItemWithValue 根据IP删除以前的旧记录
func (this *IPItemDAO) DeleteOldItemWithValue(tx *dbs.Tx, listId int64, value string) error {
	if len(value) == 0 {
		return nil
	}
	ones, err := this.Query(tx).
		ResultPk().
		UseIndex("ipFrom").
		Attr("listId", listId).
		Attr("value", value).
		Attr("state", IPItemStateEnabled).
		FindAll()
	if err != nil {
		return err
	}

	for _, one := range ones {
		var itemId = int64(one.(*IPItem).Id)
		version, err := SharedIPListDAO.IncreaseVersion(tx)
		if err != nil {
			return err
		}

		err = this.Query(tx).
			Pk(itemId).
			Set("version", version).
			Set("state", IPItemStateDisabled).
			UpdateQuickly()
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateIPItem 创建IP
func (this *IPItemDAO) CreateIPItem(tx *dbs.Tx,
	listId int64,
	value string,
	ipFrom string,
	ipTo string,
	expiredAt int64,
	reason string,
	itemType IPItemType,
	eventLevel string,
	nodeId int64,
	serverId int64,
	sourceNodeId int64,
	sourceServerId int64,
	sourceHTTPFirewallPolicyId int64,
	sourceHTTPFirewallRuleGroupId int64,
	sourceHTTPFirewallRuleSetId int64,
	shouldNotify bool) (int64, error) {
	// generate 'itemType'
	if itemType != IPItemTypeAll && len(ipFrom) > 0 {
		if iputils.IsIPv4(ipFrom) {
			itemType = IPItemTypeIPv4
		} else if iputils.IsIPv6(ipFrom) {
			itemType = IPItemTypeIPv6
		}
	}

	version, err := SharedIPListDAO.IncreaseVersion(tx)
	if err != nil {
		return 0, err
	}

	var op = NewIPItemOperator()
	op.ListId = listId
	op.Value = value
	op.IpFrom = ipFrom
	op.IpTo = ipTo

	op.Reason = reason
	op.Type = itemType
	op.EventLevel = eventLevel
	op.Version = version
	if expiredAt < 0 {
		expiredAt = 0
	}
	op.ExpiredAt = expiredAt

	op.NodeId = nodeId
	op.ServerId = serverId
	op.SourceNodeId = sourceNodeId
	op.SourceServerId = sourceServerId
	op.SourceHTTPFirewallPolicyId = sourceHTTPFirewallPolicyId
	op.SourceHTTPFirewallRuleGroupId = sourceHTTPFirewallRuleGroupId
	op.SourceHTTPFirewallRuleSetId = sourceHTTPFirewallRuleSetId

	// 服务所属用户
	if sourceServerId > 0 {
		userId, err := SharedServerDAO.FindServerUserId(tx, sourceServerId)
		if err != nil {
			return 0, err
		}
		op.SourceUserId = userId
	}

	var autoAdded = firewallconfigs.IsGlobalListId(listId) || sourceNodeId > 0 || sourceServerId > 0 || sourceHTTPFirewallPolicyId > 0
	if autoAdded {
		op.IsRead = 0
	}

	op.State = IPItemStateEnabled
	op.UpdatedAt = time.Now().Unix()

	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	itemId := types.Int64(op.Id)

	// 自动加入名单不需要即时更新，防止数量过多而导致性能问题
	if autoAdded {
		return itemId, nil
	}

	if shouldNotify {
		err = this.NotifyUpdate(tx, itemId)
		if err != nil {
			return 0, err
		}
	}
	return itemId, nil
}

// UpdateIPItem 修改IP
func (this *IPItemDAO) UpdateIPItem(tx *dbs.Tx, itemId int64, value string, ipFrom string, ipTo string, expiredAt int64, reason string, itemType IPItemType, eventLevel string) error {
	if itemId <= 0 {
		return errors.New("invalid itemId")
	}

	// generate 'itemType'
	if itemType != IPItemTypeAll && len(ipFrom) > 0 {
		if iputils.IsIPv4(ipFrom) {
			itemType = IPItemTypeIPv4
		} else if iputils.IsIPv6(ipFrom) {
			itemType = IPItemTypeIPv6
		}
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

	var op = NewIPItemOperator()
	op.Id = itemId
	op.Value = value
	op.IpFrom = ipFrom
	op.IpTo = ipTo

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
func (this *IPItemDAO) CountIPItemsWithListId(tx *dbs.Tx, listId int64, sourceUserId int64, keyword string, ipFrom string, ipTo string, eventLevel string) (int64, error) {
	var query = this.Query(tx).
		State(IPItemStateEnabled).
		Attr("listId", listId)
	if sourceUserId > 0 {
		if listId <= 0 || firewallconfigs.IsGlobalListId(listId) {
			query.Attr("sourceUserId", sourceUserId)
		}
	}
	if len(keyword) > 0 {
		query.Where("(ipFrom LIKE :keyword OR ipTo LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if len(ipFrom) > 0 {
		query.Attr("ipFrom", ipFrom)
	}
	if len(ipTo) > 0 {
		query.Attr("ipTo", ipTo)
	}
	if len(eventLevel) > 0 {
		query.Attr("eventLevel", eventLevel)
	}
	return query.Count()
}

// ListIPItemsWithListId 查找IP列表
func (this *IPItemDAO) ListIPItemsWithListId(tx *dbs.Tx, listId int64, sourceUserId int64, keyword string, ipFrom string, ipTo string, eventLevel string, offset int64, size int64) (result []*IPItem, err error) {
	var query = this.Query(tx).
		State(IPItemStateEnabled).
		Attr("listId", listId)
	if sourceUserId > 0 {
		if listId <= 0 || firewallconfigs.IsGlobalListId(listId) {
			query.Attr("sourceUserId", sourceUserId)
		}
	}
	if len(keyword) > 0 {
		query.Where("(ipFrom LIKE :keyword OR ipTo LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if len(ipFrom) > 0 {
		query.Attr("ipFrom", ipFrom)
	}
	if len(ipTo) > 0 {
		query.Attr("ipTo", ipTo)
	}
	if len(eventLevel) > 0 {
		query.Attr("eventLevel", eventLevel)
	}
	_, err = query.
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
		UseIndex("version").
		// 这里不要设置状态参数，因为我们要知道哪些是删除的
		Gt("version", version).
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
func (this *IPItemDAO) FindEnabledItemContainsIP(tx *dbs.Tx, listId int64, ip string) (*IPItem, error) {
	var query = this.Query(tx).
		Attr("listId", listId).
		State(IPItemStateEnabled)

	if iputils.IsIPv4(ip) {
		query.Where("(type='all' OR ipFrom =:ip OR INET_ATON(:ip) BETWEEN INET_ATON(ipFrom) AND INET_ATON(ipTo))").
			Param("ip", ip)
	} else if iputils.IsIPv6(ip) {
		query.Where("(type='all' OR ipFrom =:ip OR HEX(INET6_ATON(:ip)) BETWEEN HEX(INET6_ATON(ipFrom)) AND HEX(INET6_ATON(ipTo)))").
			Param("ip", ip)
	} else {
		return nil, nil
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

// FindEnabledItemsWithIP 根据IP查找Item
func (this *IPItemDAO) FindEnabledItemsWithIP(tx *dbs.Tx, ip string) (result []*IPItem, err error) {
	_, err = this.Query(tx).
		State(IPItemStateEnabled).
		Attr("ipFrom", ip).
		Attr("ipTo", "").
		Where("(expiredAt=0 OR expiredAt>:nowTime)").
		Param("nowTime", time.Now().Unix()).
		Where("listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE state=1)").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// ExistsEnabledItem 检查IP是否存在
func (this *IPItemDAO) ExistsEnabledItem(tx *dbs.Tx, itemId int64) (bool, error) {
	return this.Query(tx).
		Pk(itemId).
		State(IPItemStateEnabled).
		Exist()
}

// CountAllEnabledIPItems 计算数量
func (this *IPItemDAO) CountAllEnabledIPItems(tx *dbs.Tx, sourceUserId int64, keyword string, ip string, listId int64, unread bool, eventLevel string, listType string, isGlobal bool) (int64, error) {
	var query = this.Query(tx)
	var globalListIdStrings = strings.Join(firewallconfigs.FindGlobalListIdStrings(), ",")
	if len(listType) > 0 {
		var globalListId = firewallconfigs.FindGlobalListIdWithType(listType)
		if globalListId > 0 {
			globalListIdStrings = types.String(globalListId)
		}
	}

	if sourceUserId > 0 {
		if listId <= 0 {
			if isGlobal {
				query.Where("(listId IN (" + globalListIdStrings + ") AND sourceUserId=:sourceUserId)")
			} else {
				query.Where("((listId IN (" + globalListIdStrings + ") AND sourceUserId=:sourceUserId) OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE userId=:sourceUserId AND state=1))")
			}
			query.Param("sourceUserId", sourceUserId)
		} else if firewallconfigs.IsGlobalListId(listId) {
			query.Attr("sourceUserId", sourceUserId)
			query.UseIndex("sourceUserId")
		}
	}
	if len(keyword) > 0 {
		if net.ParseIP(keyword) != nil { // 是一个IP地址
			if iputils.IsIPv4(keyword) {
				query.Where("(type='all' OR ipFrom =:ipKeyword OR INET_ATON(:ipKeyword) BETWEEN INET_ATON(ipFrom) AND INET_ATON(ipTo))").
					Param("ipKeyword", keyword)
			} else if iputils.IsIPv6(keyword) {
				query.Where("(type='all' OR ipFrom =:ipKeyword OR HEX(INET6_ATON(:ipKeyword)) BETWEEN HEX(INET6_ATON(ipFrom)) AND HEX(INET6_ATON(ipTo)))").
					Param("ipKeyword", keyword)
			}
		} else {
			query.Like("ipFrom", dbutils.QuoteLike(keyword))
		}
	}
	if len(ip) > 0 {
		query.Attr("ipFrom", ip)
	}
	if listId > 0 {
		query.Attr("listId", listId)
	} else {
		if len(listType) > 0 {
			if isGlobal {
				query.Where("(listId IN (" + globalListIdStrings + "))")
			} else {
				query.Where("(listId IN (" + globalListIdStrings + ") OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE state=1 AND type=:listType))")
			}
			query.Param("listType", listType)
		} else {
			if isGlobal {
				query.Where("(listId IN (" + globalListIdStrings + "))")
			} else {
				query.Where("(listId IN (" + globalListIdStrings + ") OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE state=1))")
			}
		}
	}
	if unread {
		query.Attr("isRead", 0)
	}
	if len(eventLevel) > 0 {
		query.Attr("eventLevel", eventLevel)
	}

	return query.
		State(IPItemStateEnabled).
		Where("(expiredAt=0 OR expiredAt>:expiredAt)").
		Param("expiredAt", time.Now().Unix()).
		Count()
}

// ListAllEnabledIPItems 搜索所有IP
func (this *IPItemDAO) ListAllEnabledIPItems(tx *dbs.Tx, sourceUserId int64, keyword string, ip string, listId int64, unread bool, eventLevel string, listType string, isGlobal bool, offset int64, size int64) (result []*IPItem, err error) {
	var globalListIdStrings = strings.Join(firewallconfigs.FindGlobalListIdStrings(), ",")
	if len(listType) > 0 {
		var globalListId = firewallconfigs.FindGlobalListIdWithType(listType)
		if globalListId > 0 {
			globalListIdStrings = types.String(globalListId)
		}
	}

	var query = this.Query(tx)
	if sourceUserId > 0 {
		if listId <= 0 {
			if isGlobal {
				query.Where("(listId IN (" + globalListIdStrings + ") AND sourceUserId=:sourceUserId)")
			} else {
				query.Where("((listId IN (" + globalListIdStrings + ") AND sourceUserId=:sourceUserId) OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE userId=:sourceUserId AND state=1))")
			}
			query.Param("sourceUserId", sourceUserId)
		} else if firewallconfigs.IsGlobalListId(listId) {
			query.Attr("sourceUserId", sourceUserId)
			query.UseIndex("sourceUserId")
		}
	}
	if len(keyword) > 0 {
		if net.ParseIP(keyword) != nil { // 是一个IP地址
			if iputils.IsIPv4(keyword) {
				query.Where("(type='all' OR ipFrom =:ipKeyword OR INET_ATON(:ipKeyword) BETWEEN INET_ATON(ipFrom) AND INET_ATON(ipTo))").
					Param("ipKeyword", keyword)
			} else if iputils.IsIPv6(keyword) {
				query.Where("(type='all' OR ipFrom =:ipKeyword OR HEX(INET6_ATON(:ipKeyword)) BETWEEN HEX(INET6_ATON(ipFrom)) AND HEX(INET6_ATON(ipTo)))").
					Param("ipKeyword", keyword)
			}
		} else {
			query.Like("ipFrom", dbutils.QuoteLike(keyword))
		}
	}
	if len(ip) > 0 {
		query.Attr("ipFrom", ip)
	}
	if listId > 0 {
		query.Attr("listId", listId)
	} else {
		if len(listType) > 0 {
			if isGlobal {
				query.Where("(listId IN (" + globalListIdStrings + "))")
			} else {
				query.Where("(listId IN (" + globalListIdStrings + ") OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE state=1 AND type=:listType))")
			}
			query.Param("listType", listType)
		} else {
			if isGlobal {
				query.Where("(listId IN (" + globalListIdStrings + "))")
			} else {
				query.Where("(listId IN (" + globalListIdStrings + ") OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE state=1))")
			}
		}
	}
	if unread {
		query.Attr("isRead", 0)
	}
	if len(eventLevel) > 0 {
		query.Attr("eventLevel", eventLevel)
	}
	_, err = query.
		State(IPItemStateEnabled).
		Where("(expiredAt=0 OR expiredAt>:expiredAt)").
		Param("expiredAt", time.Now().Unix()).
		DescPk().
		Offset(offset).
		Size(size).
		Slice(&result).
		FindAll()
	return
}

// ListAllIPItemIds 搜索所有IP Id列表
func (this *IPItemDAO) ListAllIPItemIds(tx *dbs.Tx, sourceUserId int64, keyword string, ip string, listId int64, unread bool, eventLevel string, listType string, offset int64, size int64) (itemIds []int64, err error) {
	var globalListIdStrings = strings.Join(firewallconfigs.FindGlobalListIdStrings(), ",")
	if len(listType) > 0 {
		var globalListId = firewallconfigs.FindGlobalListIdWithType(listType)
		if globalListId > 0 {
			globalListIdStrings = types.String(globalListId)
		}
	}

	var query = this.Query(tx)
	if sourceUserId > 0 {
		if listId <= 0 {
			query.Where("((listId IN (" + globalListIdStrings + ") AND sourceUserId=:sourceUserId) OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE userId=:sourceUserId AND state=1))")
			query.Param("sourceUserId", sourceUserId)
		} else if firewallconfigs.IsGlobalListId(listId) {
			query.Attr("sourceUserId", sourceUserId)
			query.UseIndex("sourceUserId")
		}
	}
	if len(keyword) > 0 {
		if net.ParseIP(keyword) != nil { // 是一个IP地址
			query.Attr("ipFrom", keyword)
		} else {
			query.Like("ipFrom", dbutils.QuoteLike(keyword))
		}
	}
	if len(ip) > 0 {
		query.Attr("ipFrom", ip)
	}
	if listId > 0 {
		query.Attr("listId", listId)
	} else {
		if len(listType) > 0 {
			query.Where("(listId IN (" + globalListIdStrings + ") OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE state=1 AND type=:listType))")
			query.Param("listType", listType)
		} else {
			query.Where("(listId IN (" + globalListIdStrings + ") OR listId IN (SELECT id FROM " + SharedIPListDAO.Table + " WHERE state=1))")
		}
	}
	if unread {
		query.Attr("isRead", 0)
	}
	if len(eventLevel) > 0 {
		query.Attr("eventLevel", eventLevel)
	}
	result, err := query.
		ResultPk().
		State(IPItemStateEnabled).
		Where("(expiredAt=0 OR expiredAt>:expiredAt)").
		Param("expiredAt", time.Now().Unix()).
		DescPk().
		Offset(offset).
		Size(size).
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, itemOne := range result {
		itemIds = append(itemIds, int64(itemOne.(*IPItem).Id))
	}
	return
}

// UpdateItemsRead 设置所有未已读
func (this *IPItemDAO) UpdateItemsRead(tx *dbs.Tx, sourceUserId int64) error {
	var query = this.Query(tx).
		Attr("isRead", 0).
		Set("isRead", 1)

	if sourceUserId > 0 {
		query.Attr("sourceUserId", sourceUserId)
		query.UseIndex("sourceUserId")
	}

	return query.UpdateQuickly()
}

// CleanExpiredIPItems 清除过期数据
func (this *IPItemDAO) CleanExpiredIPItems(tx *dbs.Tx) error {
	// 删除 N 天之前过期的数据
	_, err := this.Query(tx).
		Where("(createdAt<=:timestamp AND updatedAt<=:timestamp)").
		State(IPItemStateDisabled).
		Param("timestamp", time.Now().Unix()-7*86400). // N 天之前过期的
		Limit(10000).                                  // 限制条数，防止数量过多导致超时
		Delete()
	if err != nil {
		return err
	}

	// 将过期的设置为已删除，这样是为了在 expiredAt<UNIX_TIMESTAMP()边缘节点让过期的IP有一个执行删除的机会
	ones, _, err := this.Query(tx).
		ResultPk().
		Where("(expiredAt>0 AND expiredAt<=:timestamp)").
		Param("timestamp", time.Now().Unix()).
		State(IPItemStateEnabled).
		Limit(500).
		FindOnes()
	if err != nil {
		return err
	}
	for _, one := range ones {
		var expiredId = one.GetInt64("id")
		newVersion, err := SharedIPListDAO.IncreaseVersion(tx)
		if err != nil {
			return err
		}
		// 这里不重置过期时间用于清理
		_, err = this.Query(tx).
			Pk(expiredId).
			Set("state", IPItemStateDisabled).
			Set("version", newVersion).
			Update()

		if err != nil {
			return err
		}
	}

	return nil
}

// ParseIPValue 解析IP值
func (this *IPItemDAO) ParseIPValue(value string) (newValue string, ipFrom string, ipTo string, ok bool) {
	if len(value) == 0 {
		return
	}

	newValue = value

	// ip1-ip2
	if strings.Contains(value, "-") {
		var pieces = strings.Split(value, "-")
		if len(pieces) != 2 {
			return
		}

		ipFrom = strings.TrimSpace(pieces[0])
		ipTo = strings.TrimSpace(pieces[1])

		if !iputils.IsValid(ipFrom) || !iputils.IsValid(ipTo) {
			return
		}

		if !iputils.IsSameVersion(ipFrom, ipTo) {
			return
		}

		if iputils.CompareIP(ipFrom, ipTo) > 0 {
			ipFrom, ipTo = ipTo, ipFrom
			newValue = ipFrom + "-" + ipTo
		}

		ok = true
		return
	}

	// ip/mask
	if strings.Contains(value, "/") {
		cidr, err := iputils.ParseCIDR(value)
		if err != nil {
			return
		}
		return newValue, cidr.From().String(), cidr.To().String(), true
	}

	// single value
	if iputils.IsValid(value) {
		ipFrom = value
		ok = true
		return
	}

	return
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

	if firewallconfigs.IsGlobalListId(listId) {
		sourceNodeId, err := this.Query(tx).
			Pk(itemId).
			Result("sourceNodeId").
			FindInt64Col(0)
		if err != nil {
			return err
		}
		if sourceNodeId > 0 {
			clusterIds, err := SharedNodeDAO.FindEnabledNodeClusterIds(tx, sourceNodeId)
			if err != nil {
				return err
			}
			for _, clusterId := range clusterIds {
				err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, 0, NodeTaskTypeIPItemChanged)
				if err != nil {
					return err
				}
			}
		} else {
			clusterIds, err := SharedNodeClusterDAO.FindAllEnabledNodeClusterIds(tx)
			if err != nil {
				return err
			}
			for _, clusterId := range clusterIds {
				err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, 0, NodeTaskTypeIPItemChanged)
				if err != nil {
					return err
				}
			}
		}
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
			err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, 0, NodeTaskTypeIPItemChanged)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
