package models

import (
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ipconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
)

const (
	IPListStateEnabled  = 1 // 已启用
	IPListStateDisabled = 0 // 已禁用
)

var listTypeCacheMap = map[int64]*IPList{} // listId => *IPList
var DefaultGlobalIPList = &IPList{
	Id:       uint32(firewallconfigs.GlobalListId),
	Name:     "全局封锁名单",
	IsPublic: true,
	IsGlobal: true,
	Type:     "black",
	State:    IPListStateEnabled,
	IsOn:     true,
}

type IPListDAO dbs.DAO

func NewIPListDAO() *IPListDAO {
	return dbs.NewDAO(&IPListDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeIPLists",
			Model:  new(IPList),
			PkName: "id",
		},
	}).(*IPListDAO)
}

var SharedIPListDAO *IPListDAO

func init() {
	dbs.OnReady(func() {
		SharedIPListDAO = NewIPListDAO()
	})
}

// EnableIPList 启用条目
func (this *IPListDAO) EnableIPList(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPListStateEnabled).
		Update()
	return err
}

// DisableIPList 禁用条目
func (this *IPListDAO) DisableIPList(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPListStateDisabled).
		Update()
	return err
}

// FindEnabledIPList 查找启用中的条目
func (this *IPListDAO) FindEnabledIPList(tx *dbs.Tx, id int64, cacheMap *utils.CacheMap) (*IPList, error) {
	if id == firewallconfigs.GlobalListId {
		return DefaultGlobalIPList, nil
	}

	var cacheKey = this.Table + ":FindEnabledIPList:" + types.String(id)
	if cacheMap != nil {
		cache, ok := cacheMap.Get(cacheKey)
		if ok {
			return cache.(*IPList), nil
		}
	}

	result, err := this.Query(tx).
		Pk(id).
		Attr("state", IPListStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, result)
	}

	return result.(*IPList), err
}

// FindIPListName 根据主键查找名称
func (this *IPListDAO) FindIPListName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindIPListCacheable 获取名单
func (this *IPListDAO) FindIPListCacheable(tx *dbs.Tx, listId int64) (*IPList, error) {
	// 全局黑名单
	if listId == firewallconfigs.GlobalListId {
		return DefaultGlobalIPList, nil
	}

	// 检查缓存
	SharedCacheLocker.RLock()
	list, ok := listTypeCacheMap[listId]
	SharedCacheLocker.RUnlock()
	if ok {
		return list, nil
	}

	one, err := this.Query(tx).
		Pk(listId).
		Result("isGlobal", "type", "state", "id", "isPublic", "isGlobal").
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	// 保存缓存
	SharedCacheLocker.Lock()
	listTypeCacheMap[listId] = one.(*IPList)
	SharedCacheLocker.Unlock()

	return one.(*IPList), nil
}

// CreateIPList 创建名单
func (this *IPListDAO) CreateIPList(tx *dbs.Tx, userId int64, serverId int64, listType ipconfigs.IPListType, name string, code string, timeoutJSON []byte, description string, isPublic bool, isGlobal bool) (int64, error) {
	var op = NewIPListOperator()
	op.IsOn = true
	op.UserId = userId
	op.ServerId = serverId
	op.State = IPListStateEnabled
	op.Type = listType
	op.Name = name
	op.Code = code
	if len(timeoutJSON) > 0 {
		op.Timeout = timeoutJSON
	}
	op.Description = description
	op.IsPublic = isPublic
	op.IsGlobal = isGlobal
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateIPList 修改名单
func (this *IPListDAO) UpdateIPList(tx *dbs.Tx, listId int64, name string, code string, timeoutJSON []byte, description string) error {
	if listId <= 0 {
		return errors.New("invalid listId")
	}
	op := NewIPListOperator()
	op.Id = listId
	op.Name = name
	op.Code = code
	if len(timeoutJSON) > 0 {
		op.Timeout = timeoutJSON
	} else {
		op.Timeout = "null"
	}
	op.Description = description
	err := this.Save(tx, op)
	return err
}

// IncreaseVersion 增加版本
func (this *IPListDAO) IncreaseVersion(tx *dbs.Tx) (int64, error) {
	return SharedSysLockerDAO.Increase(tx, "IP_LIST_VERSION", 1000000)
}

// CheckUserIPList 检查用户权限
func (this *IPListDAO) CheckUserIPList(tx *dbs.Tx, userId int64, listId int64) error {
	if userId == 0 || listId == 0 {
		return ErrNotFound
	}

	// 获取名单信息
	listOne, err := this.Query(tx).
		Pk(listId).
		Result("userId", "serverId").
		Find()
	if err != nil {
		return err
	}
	if listOne == nil {
		return ErrNotFound
	}
	var list = listOne.(*IPList)
	if int64(list.UserId) == userId {
		return nil
	}

	var serverId = int64(list.ServerId)
	if serverId > 0 {
		return SharedServerDAO.CheckUserServer(tx, userId, serverId)
	}

	return ErrNotFound
}

// CountAllEnabledIPLists 计算名单数量
func (this *IPListDAO) CountAllEnabledIPLists(tx *dbs.Tx, listType string, isPublic bool, keyword string) (int64, error) {
	var query = this.Query(tx).
		State(IPListStateEnabled).
		Attr("type", listType).
		Attr("isPublic", isPublic)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	return query.Count()
}

// ListEnabledIPLists 列出单页名单
func (this *IPListDAO) ListEnabledIPLists(tx *dbs.Tx, listType string, isPublic bool, keyword string, offset int64, size int64) (result []*IPList, err error) {
	var query = this.Query(tx).
		State(IPListStateEnabled).
		Attr("type", listType).
		Attr("isPublic", isPublic)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	_, err = query.Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// ExistsEnabledIPList 检查IP名单是否存在
func (this *IPListDAO) ExistsEnabledIPList(tx *dbs.Tx, listId int64) (bool, error) {
	if listId <= 0 {
		return false, nil
	}
	return this.Query(tx).
		Pk(listId).
		State(IPListStateEnabled).
		Exist()
}

// NotifyUpdate 通知更新
func (this *IPListDAO) NotifyUpdate(tx *dbs.Tx, listId int64, taskType NodeTaskType) error {
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
			err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, taskType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
