package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/sizes"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strconv"
	"strings"
	"time"
)

const (
	NodeStateEnabled  = 1 // 已启用
	NodeStateDisabled = 0 // 已禁用
)

var nodeIdCacheMap = map[string]int64{} // uniqueId => nodeId

type NodeDAO dbs.DAO

func NewNodeDAO() *NodeDAO {
	return dbs.NewDAO(&NodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodes",
			Model:  new(Node),
			PkName: "id",
		},
	}).(*NodeDAO)
}

var SharedNodeDAO *NodeDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeDAO = NewNodeDAO()
	})
}

// EnableNode 启用条目
func (this *NodeDAO) EnableNode(tx *dbs.Tx, id uint32) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", NodeStateEnabled).
		Update()
}

// DisableNode 禁用条目
func (this *NodeDAO) DisableNode(tx *dbs.Tx, nodeId int64) (err error) {
	// 删除缓存
	uniqueId, err := this.Query(tx).
		Pk(nodeId).
		Result("uniqueId").
		FindStringCol("")
	if err != nil {
		return err
	}
	if len(uniqueId) > 0 {
		SharedCacheLocker.Lock()
		delete(nodeIdCacheMap, uniqueId)
		SharedCacheLocker.Unlock()
	}

	_, err = this.Query(tx).
		Pk(nodeId).
		Set("state", NodeStateDisabled).
		Update()
	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, nodeId)
	if err != nil {
		return err
	}

	err = this.NotifyDNSUpdate(tx, nodeId)
	if err != nil {
		return err
	}

	// 删除运行日志
	return SharedNodeLogDAO.DeleteNodeLogs(tx, nodeconfigs.NodeRoleNode, nodeId)
}

// FindEnabledNode 查找启用中的条目
func (this *NodeDAO) FindEnabledNode(tx *dbs.Tx, id int64) (*Node, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Node), err
}

// FindEnabledBasicNode 获取节点的基本信息
func (this *NodeDAO) FindEnabledBasicNode(tx *dbs.Tx, nodeId int64) (*Node, error) {
	one, err := this.Query(tx).
		State(NodeStateEnabled).
		Pk(nodeId).
		Result("id", "name", "clusterId", "groupId", "isOn", "isUp").
		Find()
	if one == nil {
		return nil, err
	}
	return one.(*Node), nil
}

// FindNodeName 根据主键查找名称
func (this *NodeDAO) FindNodeName(tx *dbs.Tx, id int64) (string, error) {
	name, err := this.Query(tx).
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}

// CreateNode 创建节点
func (this *NodeDAO) CreateNode(tx *dbs.Tx, adminId int64, name string, clusterId int64, groupId int64, regionId int64) (nodeId int64, err error) {
	// 检查节点数量
	if teaconst.MaxNodes > 0 {
		count, err := this.Query(tx).
			State(NodeStateEnabled).
			Where("clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)").
			Count()
		if err != nil {
			return 0, err
		}
		if int64(teaconst.MaxNodes) <= count {
			return 0, errors.New("[企业版]超出最大节点数限制：" + types.String(teaconst.MaxNodes) + "，请购买更多配额")
		}
	}

	uniqueId, err := this.GenUniqueId(tx)
	if err != nil {
		return 0, err
	}

	secret := rands.String(32)

	// 保存API Token
	err = SharedApiTokenDAO.CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleNode)
	if err != nil {
		return
	}

	var op = NewNodeOperator()
	op.AdminId = adminId
	op.Name = name
	op.UniqueId = uniqueId
	op.Secret = secret
	op.ClusterId = clusterId
	op.GroupId = groupId
	op.RegionId = regionId
	op.IsOn = 1
	op.State = NodeStateEnabled
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
func (this *NodeDAO) UpdateNode(tx *dbs.Tx, nodeId int64, name string, clusterId int64, secondaryClusterIds []int64, groupId int64, regionId int64, isOn bool, level int, lnAddrs []string) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	// 老的集群
	oldClusterIds, err := this.FindEnabledNodeClusterIds(tx, nodeId)
	if err != nil {
		return err
	}

	// 老的级别
	oldLevel, err := this.Query(tx).
		Pk(nodeId).
		Result("level").
		FindIntCol(0)
	if err != nil {
		return err
	}

	var op = NewNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.ClusterId = clusterId

	// 去重
	var filteredSecondaryClusterIds = []int64{}
	for _, secondaryClusterId := range secondaryClusterIds {
		if secondaryClusterId <= 0 {
			continue
		}
		if lists.ContainsInt64(filteredSecondaryClusterIds, secondaryClusterId) {
			continue
		}
		filteredSecondaryClusterIds = append(filteredSecondaryClusterIds, secondaryClusterId)
	}
	filteredSecondaryClusterIdsJSON, err := json.Marshal(filteredSecondaryClusterIds)
	if err != nil {
		return err
	}
	op.SecondaryClusterIds = filteredSecondaryClusterIdsJSON

	op.GroupId = groupId
	op.RegionId = regionId
	op.LatestVersion = dbs.SQL("latestVersion+1")
	op.IsOn = isOn

	if teaconst.IsPlus {
		op.Level = level

		if lnAddrs == nil {
			lnAddrs = []string{}
		}
		lnAddrsJSON, err := json.Marshal(lnAddrs)
		if err != nil {
			return err
		}
		op.LnAddrs = lnAddrsJSON
	}

	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, nodeId)
	if err != nil {
		return err
	}

	// 通知老的集群DNS更新
	for _, oldClusterId := range oldClusterIds {
		if oldClusterId != clusterId && !lists.ContainsInt64(secondaryClusterIds, oldClusterId) {
			err = dns.SharedDNSTaskDAO.CreateClusterTask(tx, oldClusterId, dns.DNSTaskTypeClusterChange)
			if err != nil {
				return err
			}
		}
	}

	// 通知子级级别变更
	if oldLevel > 1 || level > 1 {
		err = this.NotifyLevelUpdate(tx, nodeId)
		if err != nil {
			return err
		}
	}

	return this.NotifyDNSUpdate(tx, nodeId)
}

// CountAllEnabledNodes 计算所有节点数量
func (this *NodeDAO) CountAllEnabledNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Where("clusterId IN (SELECT id FROM "+SharedNodeClusterDAO.Table+" WHERE state=:clusterState)").
		Param("clusterState", NodeClusterStateEnabled).
		Count()
}

// CountAllEnabledOfflineNodes 计算所有离线节点数量
func (this *NodeDAO) CountAllEnabledOfflineNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Where("clusterId IN (SELECT id FROM "+SharedNodeClusterDAO.Table+" WHERE state=:clusterState)").
		Param("clusterState", NodeClusterStateEnabled).
		Where("(status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>60)").
		Count()
}

// ListEnabledNodesMatch 列出单页节点
func (this *NodeDAO) ListEnabledNodesMatch(tx *dbs.Tx,
	clusterId int64,
	installState configutils.BoolState,
	activeState configutils.BoolState,
	keyword string,
	groupId int64,
	regionId int64,
	level int32,
	includeSecondaryNodes bool,
	order string,
	offset int64,
	size int64) (result []*Node, err error) {
	query := this.Query(tx).
		Result(this.Table + ".*"). // must have table name for table joins below
		State(NodeStateEnabled).
		Offset(offset).
		Limit(size).
		Slice(&result)

	// 集群
	if clusterId > 0 {
		if includeSecondaryNodes {
			query.Where("("+this.Table+".clusterId=:primaryClusterId OR JSON_CONTAINS(secondaryClusterIds, :primaryClusterIdString))").
				Param("primaryClusterId", clusterId).
				Param("primaryClusterIdString", types.String(clusterId))
		} else {
			query.Attr(this.Table+".clusterId", clusterId)
		}
	} else {
		query.Where(this.Table + ".clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)")
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
		query.Where("isActive AND UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')<=60")
	case configutils.BoolStateNo:
		query.Where("(status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>60)")
	}

	// 关键词
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR JSON_EXTRACT(status,'$.hostname') LIKE :keyword OR id IN (SELECT nodeId FROM "+SharedNodeIPAddressDAO.Table+" WHERE ip LIKE :keyword))").
			Param("keyword", dbutils.QuoteLike(keyword))
	}

	// 分组
	if groupId > 0 {
		query.Attr("groupId", groupId)
	} else if groupId < 0 {
		query.Attr("groupId", 0)
	}

	// 区域
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}

	// 级别
	if level > 0 {
		query.Attr("level", level)
	}

	// 排序
	var minute = timeutil.FormatTime("YmdHi", time.Now().Unix()-60)
	var nodeValueTable = SharedNodeValueDAO.Table
	var valueItem = ""
	var valueField = ""
	var isAsc = false
	var ifNullValue int64 = 0
	switch order {
	case "cpuAsc":
		valueItem = "cpu"
		valueField = "usage"
		isAsc = true
		ifNullValue = 100
	case "cpuDesc":
		valueItem = "cpu"
		valueField = "usage"
		isAsc = false
		ifNullValue = -1
	case "memoryAsc":
		valueItem = "memory"
		valueField = "usage"
		isAsc = true
		ifNullValue = 100
	case "memoryDesc":
		valueItem = "memory"
		valueField = "usage"
		isAsc = false
		ifNullValue = -1
	case "trafficInAsc":
		valueItem = "trafficIn"
		valueField = "total"
		isAsc = true
		ifNullValue = 60 * sizes.G
	case "trafficInDesc":
		valueItem = "trafficIn"
		valueField = "total"
		isAsc = false
		ifNullValue = -1
	case "trafficOutAsc":
		valueItem = "trafficOut"
		valueField = "total"
		isAsc = true
		ifNullValue = sizes.G
	case "trafficOutDesc":
		valueItem = "trafficOut"
		valueField = "total"
		isAsc = false
		ifNullValue = -1
	case "loadAsc":
		valueItem = "load"
		valueField = "load1m"
		isAsc = true
		ifNullValue = 1000
	case "loadDesc":
		valueItem = "load"
		valueField = "load1m"
		isAsc = false
		ifNullValue = -1
	default:
		query.Desc("level")
	}

	if len(valueItem) > 0 {
		query.Join(SharedNodeValueDAO, dbs.QueryJoinLeft, this.Table+".id="+nodeValueTable+".nodeId AND "+nodeValueTable+".role='"+nodeconfigs.NodeRoleNode+"' AND "+nodeValueTable+".item='"+valueItem+"' "+" AND "+nodeValueTable+".minute=:minute")
		query.Param("minute", minute)
		var ifNullSQL = "IFNULL(CAST(JSON_EXTRACT(value, '$." + valueField + "') AS DECIMAL (20, 6)), " + types.String(ifNullValue) + ")"
		if isAsc {
			query.Asc(ifNullSQL)
		} else {
			query.Desc(ifNullSQL)
		}
	}

	query.DescPk()

	_, err = query.FindAll()
	return
}

// FindEnabledNodeWithUniqueIdAndSecret 根据节点ID和密钥查询节点
func (this *NodeDAO) FindEnabledNodeWithUniqueIdAndSecret(tx *dbs.Tx, uniqueId string, secret string) (*Node, error) {
	one, err := this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("secret", secret).
		State(NodeStateEnabled).
		Find()

	if one != nil {
		return one.(*Node), err
	}

	return nil, err
}

// FindEnabledNodeWithUniqueId 根据节点ID获取节点
func (this *NodeDAO) FindEnabledNodeWithUniqueId(tx *dbs.Tx, uniqueId string) (*Node, error) {
	one, err := this.Query(tx).
		Attr("uniqueId", uniqueId).
		State(NodeStateEnabled).
		Find()

	if one != nil {
		return one.(*Node), err
	}

	return nil, err
}

// FindNodeClusterId 获取节点集群ID
func (this *NodeDAO) FindNodeClusterId(tx *dbs.Tx, nodeId int64) (int64, error) {
	col, err := this.Query(tx).
		Pk(nodeId).
		Result("clusterId").
		FindCol(0)
	return types.Int64(col), err
}

// FindNodeLevel 获取节点级别
func (this *NodeDAO) FindNodeLevel(tx *dbs.Tx, nodeId int64) (int, error) {
	level, err := this.Query(tx).
		Pk(nodeId).
		Result("level").
		FindIntCol(0)
	if err != nil {
		return 0, err
	}
	if level < 1 {
		level = 1
	}
	return level, nil
}

// FindNodeLevelInfo 获取节点级别相关信息
func (this *NodeDAO) FindNodeLevelInfo(tx *dbs.Tx, nodeId int64) (*Node, error) {
	one, err := this.Query(tx).
		Pk(nodeId).
		Result("id", "clusterId", "secondaryClusterIds", "groupId", "level").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	var node = one.(*Node)
	if node.Level < 1 {
		node.Level = 1
	}
	return node, nil
}

// FindEnabledAndOnNodeClusterIds 获取节点所属所有可用而且启用的集群ID
func (this *NodeDAO) FindEnabledAndOnNodeClusterIds(tx *dbs.Tx, nodeId int64) (result []int64, err error) {
	one, err := this.Query(tx).
		Pk(nodeId).
		Result("clusterId", "secondaryClusterIds").
		Find()
	if one == nil {
		return nil, err
	}
	var clusterId = int64(one.(*Node).ClusterId)
	if clusterId > 0 {
		result = append(result, clusterId)
	}

	for _, clusterId := range one.(*Node).DecodeSecondaryClusterIds() {
		if lists.ContainsInt64(result, clusterId) {
			continue
		}

		// 检查是否启用
		isOn, err := SharedNodeClusterDAO.CheckNodeClusterIsOn(tx, clusterId)
		if err != nil {
			return nil, err
		}
		if !isOn {
			continue
		}

		result = append(result, clusterId)
	}
	return
}

// FindEnabledNodeClusterIds 获取节点所属所有可用的集群ID
func (this *NodeDAO) FindEnabledNodeClusterIds(tx *dbs.Tx, nodeId int64) (result []int64, err error) {
	one, err := this.Query(tx).
		Pk(nodeId).
		Result("clusterId", "secondaryClusterIds").
		Find()
	if one == nil {
		return nil, err
	}
	var clusterId = int64(one.(*Node).ClusterId)
	if clusterId > 0 {
		result = append(result, clusterId)
	}

	for _, secondaryClusterId := range one.(*Node).DecodeSecondaryClusterIds() {
		if lists.ContainsInt64(result, secondaryClusterId) {
			continue
		}

		result = append(result, secondaryClusterId)
	}
	return
}

// FindEnabledNodeIdsWithClusterId 查找某个集群下的所有节点IDs
func (this *NodeDAO) FindEnabledNodeIdsWithClusterId(tx *dbs.Tx, clusterId int64) ([]int64, error) {
	ones, err := this.Query(tx).
		Attr("clusterId", clusterId).
		State(NodeClusterStateEnabled).
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	var result = []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*Node).Id))
	}
	return result, nil
}

// FindEnabledNodesWithGroupIdAndLevel 查找当前分组下的某个级别的所有节点
func (this *NodeDAO) FindEnabledNodesWithGroupIdAndLevel(tx *dbs.Tx, groupId int64, level int) (result []*Node, err error) {
	if groupId <= 0 {
		return
	}
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Result("id", "clusterId", "secondaryClusterIds", "uniqueId", "secret", "lnAddrs").
		Attr("isOn", true).
		Attr("groupId", groupId).
		Attr("level", level).
		Slice(&result).
		FindAll()
	return
}

// FindEnabledNodesWithClusterIdAndLevel 查找当前集群下的某个级别的所有节点
func (this *NodeDAO) FindEnabledNodesWithClusterIdAndLevel(tx *dbs.Tx, clusterId int64, level int) (result []*Node, err error) {
	if clusterId <= 0 {
		return
	}
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Result("id", "clusterId", "secondaryClusterIds", "uniqueId", "secret").
		Attr("isOn", true).
		Where("(clusterId=:primaryClusterId OR JSON_CONTAINS(secondaryClusterIds, :primaryClusterIdString))").
		Param("primaryClusterId", clusterId).
		Param("primaryClusterIdString", types.String(clusterId)).
		Attr("groupId", 0). // 需要去掉已经有分组的
		Attr("level", level).
		Slice(&result).
		FindAll()
	return
}

// FindAllNodeIdsMatch 匹配节点并返回节点ID
func (this *NodeDAO) FindAllNodeIdsMatch(tx *dbs.Tx, clusterId int64, includeSecondaryNodes bool, isOn configutils.BoolState) (result []int64, err error) {
	var query = this.Query(tx)
	query.State(NodeStateEnabled)
	if clusterId > 0 {
		if includeSecondaryNodes {
			query.Where("(clusterId=:primaryClusterId OR JSON_CONTAINS(secondaryClusterIds, :primaryClusterIdString))").
				Param("primaryClusterId", clusterId).
				Param("primaryClusterIdString", types.String(clusterId))
		} else {
			query.Attr("clusterId", clusterId)
		}
	} else {
		query.Where("clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)")
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

// FindAllEnabledNodesWithClusterId 获取一个集群的所有节点
func (this *NodeDAO) FindAllEnabledNodesWithClusterId(tx *dbs.Tx, clusterId int64, includeSecondary bool) (result []*Node, err error) {
	var query = this.Query(tx)

	if includeSecondary {
		query.Where("(clusterId=:primaryClusterId OR JSON_CONTAINS(secondaryClusterIds, :primaryClusterIdString))").
			Param("primaryClusterId", clusterId).
			Param("primaryClusterIdString", types.String(clusterId))
	} else {
		query.Attr("clusterId", clusterId)
	}

	_, err = query.
		State(NodeStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()

	return
}

// FindEnabledAndOnNodeIdsWithClusterId 查找某个集群下的所有启用的节点IDs
func (this *NodeDAO) FindEnabledAndOnNodeIdsWithClusterId(tx *dbs.Tx, clusterId int64, includeSecondary bool) ([]int64, error) {
	var query = this.Query(tx)
	if includeSecondary {
		query.Where("(clusterId=:primaryClusterId OR JSON_CONTAINS(secondaryClusterIds, :primaryClusterIdString))").
			Param("primaryClusterId", clusterId).
			Param("primaryClusterIdString", types.String(clusterId))
	} else {
		query.Attr("clusterId", clusterId)
	}
	ones, err := query.
		Attr("isOn", true).
		State(NodeStateEnabled).
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	var result = []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*Node).Id))
	}
	return result, nil
}

// FindAllEnabledNodeIdsWithClusterId 获取一个集群的所有节点Ids
func (this *NodeDAO) FindAllEnabledNodeIdsWithClusterId(tx *dbs.Tx, clusterId int64) (result []int64, err error) {
	ones, err := this.Query(tx).
		ResultPk().
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		result = append(result, int64(one.(*Node).Id))
	}
	return
}

// FindAllInactiveNodesWithClusterId 取得一个集群离线的节点
func (this *NodeDAO) FindAllInactiveNodesWithClusterId(tx *dbs.Tx, clusterId int64) (result []*Node, err error) {
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Result("id", "name", "status").
		Attr("clusterId", clusterId).
		Attr("isOn", true).        // 只监控启用的节点
		Attr("isInstalled", true). // 只监控已经安装的节点
		Attr("isActive", false).
		Result("id", "name").
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledNodesMatch 计算节点数量
func (this *NodeDAO) CountAllEnabledNodesMatch(tx *dbs.Tx,
	clusterId int64,
	installState configutils.BoolState,
	activeState configutils.BoolState,
	keyword string,
	groupId int64,
	regionId int64,
	level int32,
	includeSecondaryNodes bool) (int64, error) {
	query := this.Query(tx)
	query.State(NodeStateEnabled)

	// 集群
	if clusterId > 0 {
		if includeSecondaryNodes {
			query.Where("(clusterId=:primaryClusterId OR JSON_CONTAINS(secondaryClusterIds, :primaryClusterIdString))").
				Param("primaryClusterId", clusterId).
				Param("primaryClusterIdString", types.String(clusterId))
		} else {
			query.Attr("clusterId", clusterId)
		}
	} else {
		query.Where("clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)")
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
		query.Where("isActive AND UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')<=60")
	case configutils.BoolStateNo:
		query.Where("(status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>60)")
	}

	// 关键词
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR JSON_EXTRACT(status,'$.hostname') LIKE :keyword OR id IN (SELECT nodeId FROM "+SharedNodeIPAddressDAO.Table+" WHERE ip LIKE :keyword))").
			Param("keyword", dbutils.QuoteLike(keyword))
	}

	// 分组
	if groupId > 0 {
		query.Attr("groupId", groupId)
	} else if groupId < 0 {
		query.Attr("groupId", 0)
	}

	// 区域
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}

	// 级别
	if level > 0 {
		query.Attr("level", level)
	}

	return query.Count()
}

// UpdateNodeStatus 更改节点状态
func (this *NodeDAO) UpdateNodeStatus(tx *dbs.Tx, nodeId int64, nodeStatus *nodeconfigs.NodeStatus) error {
	if nodeStatus == nil {
		return nil
	}

	nodeStatusJSON, err := json.Marshal(nodeStatus)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(nodeId).
		Set("isActive", true).
		Set("status", nodeStatusJSON).
		Update()
	return err
}

// FindNodeStatus 获取节点状态
func (this *NodeDAO) FindNodeStatus(tx *dbs.Tx, nodeId int64) (*nodeconfigs.NodeStatus, error) {
	statusJSONString, err := this.Query(tx).
		Pk(nodeId).
		Result("status").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(statusJSONString) == 0 {
		return nil, nil
	}

	status := &nodeconfigs.NodeStatus{}
	err = json.Unmarshal([]byte(statusJSONString), status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// UpdateNodeIsActive 更改节点在线状态
func (this *NodeDAO) UpdateNodeIsActive(tx *dbs.Tx, nodeId int64, isActive bool) error {
	b := "true"
	if !isActive {
		b = "false"
	}
	_, err := this.Query(tx).
		Pk(nodeId).
		Where("status IS NOT NULL").
		Set("status", dbs.SQL("JSON_SET(status, '$.isActive', "+b+")")).
		Update()
	return err
}

// UpdateNodeIsInstalled 设置节点安装状态
func (this *NodeDAO) UpdateNodeIsInstalled(tx *dbs.Tx, nodeId int64, isInstalled bool) error {
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("isInstalled", isInstalled).
		Set("installStatus", "null"). // 重置安装状态
		Update()
	return err
}

// FindNodeInstallStatus 查询节点的安装状态
func (this *NodeDAO) FindNodeInstallStatus(tx *dbs.Tx, nodeId int64) (*NodeInstallStatus, error) {
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

	installStatus := node.(*Node).InstallStatus
	isInstalled := node.(*Node).IsInstalled
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

// UpdateNodeInstallStatus 修改节点的安装状态
func (this *NodeDAO) UpdateNodeInstallStatus(tx *dbs.Tx, nodeId int64, status *NodeInstallStatus) error {
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

// ComposeNodeConfig 组合配置
// TODO 提升运行速度
func (this *NodeDAO) ComposeNodeConfig(tx *dbs.Tx, nodeId int64, cacheMap *utils.CacheMap) (*nodeconfigs.NodeConfig, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}

	node, err := this.FindEnabledNode(tx, nodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found '" + strconv.FormatInt(nodeId, 10) + "'")
	}

	if node.Level < 1 {
		node.Level = 1
	}

	var config = &nodeconfigs.NodeConfig{
		Id:       int64(node.Id),
		NodeId:   node.UniqueId,
		Secret:   node.Secret,
		IsOn:     node.IsOn,
		Servers:  nil,
		Version:  int64(node.Version),
		Name:     node.Name,
		MaxCPU:   types.Int32(node.MaxCPU),
		RegionId: int64(node.RegionId),
		Level:    types.Int32(node.Level),
		GroupId:  int64(node.GroupId),
	}

	// API节点IP
	apiNodeIPs, err := SharedAPINodeDAO.FindAllEnabledAPIAccessIPs(tx, cacheMap)
	if err != nil {
		return nil, err
	}
	config.AllowedIPs = append(config.AllowedIPs, apiNodeIPs...)

	// 获取所有的服务
	servers, err := SharedServerDAO.FindAllEnabledServersWithNode(tx, int64(node.Id))
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		serverConfig, err := SharedServerDAO.ComposeServerConfig(tx, server, cacheMap, true)
		if err != nil {
			return nil, err
		}
		if serverConfig == nil {
			continue
		}
		config.Servers = append(config.Servers, serverConfig)

		if server.IsOn && server.SupportCNAME == 1 {
			config.SupportCNAME = true
		}
	}

	// 全局设置
	// TODO 根据用户的不同读取不同的全局设置
	var settingCacheKey = "SharedSysSettingDAO:" + systemconfigs.SettingCodeServerGlobalConfig
	settingJSONCache, ok := cacheMap.Get(settingCacheKey)
	var settingJSON = []byte{}
	if ok {
		settingJSON = settingJSONCache.([]byte)
	} else {
		settingJSON, err = SharedSysSettingDAO.ReadSetting(tx, systemconfigs.SettingCodeServerGlobalConfig)
		if err != nil {
			return nil, err
		}
		cacheMap.Put(settingCacheKey, settingJSON)
	}

	if len(settingJSON) > 0 {
		globalConfig := &serverconfigs.GlobalConfig{}
		err = json.Unmarshal(settingJSON, globalConfig)
		if err != nil {
			return nil, err
		}
		config.GlobalConfig = globalConfig
	}

	var primaryClusterId = int64(node.ClusterId)
	var clusterIds = []int64{primaryClusterId}
	clusterIds = append(clusterIds, node.DecodeSecondaryClusterIds()...)
	var clusterIndex = 0
	config.WebPImagePolicies = map[int64]*nodeconfigs.WebPImagePolicy{}
	config.UAMPolicies = map[int64]*nodeconfigs.UAMPolicy{}
	var allowIPMaps = map[string]bool{}
	for _, clusterId := range clusterIds {
		nodeCluster, err := SharedNodeClusterDAO.FindClusterBasicInfo(tx, clusterId, cacheMap)
		if err != nil {
			return nil, err
		}
		if nodeCluster == nil || !nodeCluster.IsOn {
			continue
		}

		// 节点IP地址
		nodeIPAddresses, err := SharedNodeIPAddressDAO.FindAllAccessibleIPAddressesWithClusterId(tx, nodeconfigs.NodeRoleNode, clusterId, cacheMap)
		if err != nil {
			return nil, err
		}
		for _, address := range nodeIPAddresses {
			var ip = address.Ip
			_, ok := allowIPMaps[ip]
			if !ok {
				allowIPMaps[ip] = true
				config.AllowedIPs = append(config.AllowedIPs, ip)
			}
		}

		// 防火墙
		var httpFirewallPolicyId = int64(nodeCluster.HttpFirewallPolicyId)
		if httpFirewallPolicyId > 0 {
			firewallPolicy, err := SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, httpFirewallPolicyId, cacheMap)
			if err != nil {
				return nil, err
			}
			if firewallPolicy != nil {
				config.HTTPFirewallPolicies = append(config.HTTPFirewallPolicies, firewallPolicy)
			}
		}

		// 缓存策略
		var httpCachePolicyId = int64(nodeCluster.CachePolicyId)
		if httpCachePolicyId > 0 {
			cachePolicy, err := SharedHTTPCachePolicyDAO.ComposeCachePolicy(tx, httpCachePolicyId, cacheMap)
			if err != nil {
				return nil, err
			}
			if cachePolicy != nil {
				config.HTTPCachePolicies = append(config.HTTPCachePolicies, cachePolicy)
			}
		}

		// 时区
		if len(config.TimeZone) == 0 {
			var timeZone = nodeCluster.TimeZone
			if len(timeZone) > 0 {
				config.TimeZone = timeZone
			}
		}

		// 最大线程数、TCP连接数
		if clusterIndex == 0 {
			config.MaxThreads = int(nodeCluster.NodeMaxThreads)
			config.DDoSProtection = nodeCluster.DecodeDDoSProtection()
			config.AutoOpenPorts = nodeCluster.AutoOpenPorts == 1
		}

		// webp
		if IsNotNull(nodeCluster.Webp) {
			var webpPolicy = &nodeconfigs.WebPImagePolicy{}
			err = json.Unmarshal(nodeCluster.Webp, webpPolicy)
			if err != nil {
				return nil, err
			}
			config.WebPImagePolicies[clusterId] = webpPolicy
		}

		// UAM
		if IsNotNull(nodeCluster.Uam) {
			var uamPolicy = &nodeconfigs.UAMPolicy{}
			err = json.Unmarshal(nodeCluster.Uam, uamPolicy)
			if err != nil {
				return nil, err
			}
			config.UAMPolicies[clusterId] = uamPolicy
		}

		clusterIndex++
	}

	// 缓存最大容量设置
	if len(node.MaxCacheDiskCapacity) > 0 {
		capacity := &shared.SizeCapacity{}
		err = json.Unmarshal(node.MaxCacheDiskCapacity, capacity)
		if err != nil {
			return nil, err
		}
		if capacity.Count > 0 {
			config.MaxCacheDiskCapacity = capacity
		}
	}

	if len(node.MaxCacheMemoryCapacity) > 0 {
		capacity := &shared.SizeCapacity{}
		err = json.Unmarshal(node.MaxCacheMemoryCapacity, capacity)
		if err != nil {
			return nil, err
		}
		if capacity.Count > 0 {
			config.MaxCacheMemoryCapacity = capacity
		}
	}

	config.CacheDiskDir = node.CacheDiskDir

	// TOA
	toaConfig, err := SharedNodeClusterDAO.FindClusterTOAConfig(tx, primaryClusterId, cacheMap)
	if err != nil {
		return nil, err
	}
	config.TOA = toaConfig

	// 系统服务
	services, err := SharedNodeClusterDAO.FindNodeClusterSystemServices(tx, primaryClusterId, cacheMap)
	if err != nil {
		return nil, err
	}
	if len(services) > 0 {
		config.SystemServices = services
	}

	// DNS Resolver
	if IsNotNull(node.DnsResolver) {
		var dnsResolverConfig = nodeconfigs.DefaultDNSResolverConfig()
		err = json.Unmarshal(node.DnsResolver, dnsResolverConfig)
		if err != nil {
			return nil, err
		}
		config.DNSResolver = dnsResolverConfig
	}

	// 防火墙动作
	actions, err := SharedNodeClusterFirewallActionDAO.FindAllEnabledFirewallActions(tx, primaryClusterId, cacheMap)
	if err != nil {
		return nil, err
	}
	for _, action := range actions {
		actionConfig, err := SharedNodeClusterFirewallActionDAO.ComposeFirewallActionConfig(tx, action)
		if err != nil {
			return nil, err
		}
		if actionConfig != nil {
			config.FirewallActions = append(config.FirewallActions, actionConfig)
		}
	}

	// 集群指标
	metricItemIds, err := SharedNodeClusterMetricItemDAO.FindAllClusterItemIds(tx, int64(node.ClusterId), cacheMap)
	if err != nil {
		return nil, err
	}
	var metricItems = []*serverconfigs.MetricItemConfig{}
	for _, itemId := range metricItemIds {
		itemConfig, err := SharedMetricItemDAO.ComposeItemConfig(tx, itemId)
		if err != nil {
			return nil, err
		}
		if itemConfig != nil {
			metricItems = append(metricItems, itemConfig)
		}
	}

	// 公用指标
	publicMetricItems, err := SharedMetricItemDAO.FindAllPublicItems(tx, serverconfigs.MetricItemCategoryHTTP, cacheMap)
	if err != nil {
		return nil, err
	}
	for _, item := range publicMetricItems {
		itemConfig := SharedMetricItemDAO.ComposeItemConfigWithItem(item)
		if itemConfig != nil && !lists.ContainsInt64(metricItemIds, itemConfig.Id) {
			metricItems = append(metricItems, itemConfig)
		}
	}

	config.MetricItems = metricItems

	// 产品
	adminUIConfig, err := SharedSysSettingDAO.ReadAdminUIConfig(tx, cacheMap)
	if err != nil {
		return nil, err
	}
	if adminUIConfig != nil {
		config.ProductConfig = &nodeconfigs.ProductConfig{
			Name:    adminUIConfig.ProductName,
			Version: adminUIConfig.Version,
		}
	}

	// OCSP
	ocspVersion, err := SharedSSLCertDAO.FindCertOCSPLatestVersion(tx)
	if err != nil {
		return nil, err
	}
	config.OCSPVersion = ocspVersion

	// DDOS Protection
	var ddosProtection = node.DecodeDDoSProtection()
	if ddosProtection != nil {
		if config.DDoSProtection == nil {
			config.DDoSProtection = ddosProtection
		} else {
			config.DDoSProtection.Merge(ddosProtection)
		}
	}

	// 初始化扩展配置
	err = this.composeExtConfig(tx, config, clusterIds, cacheMap)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// UpdateNodeConnectedAPINodes 修改当前连接的API节点
func (this *NodeDAO) UpdateNodeConnectedAPINodes(tx *dbs.Tx, nodeId int64, apiNodeIds []int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	var op = NewNodeOperator()
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

// FindEnabledNodeIdWithUniqueId 根据UniqueId获取ID
func (this *NodeDAO) FindEnabledNodeIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("uniqueId", uniqueId).
		ResultPk().
		FindInt64Col(0)
}

// FindEnabledNodeIdWithUniqueIdCacheable 根据UniqueId获取ID，并可以使用缓存
func (this *NodeDAO) FindEnabledNodeIdWithUniqueIdCacheable(tx *dbs.Tx, uniqueId string) (int64, error) {
	SharedCacheLocker.RLock()
	nodeId, ok := nodeIdCacheMap[uniqueId]
	if ok {
		SharedCacheLocker.RUnlock()
		return nodeId, nil
	}
	SharedCacheLocker.RUnlock()
	nodeId, err := this.Query(tx).
		State(NodeStateEnabled).
		Attr("uniqueId", uniqueId).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}
	if nodeId > 0 {
		SharedCacheLocker.Lock()
		nodeIdCacheMap[uniqueId] = nodeId
		SharedCacheLocker.Unlock()
	}
	return nodeId, nil
}

// CountAllEnabledNodesWithGrantId 计算使用某个认证的节点数量
func (this *NodeDAO) CountAllEnabledNodesWithGrantId(tx *dbs.Tx, grantId int64) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Where("id IN (SELECT nodeId FROM edgeNodeLogins WHERE type='ssh' AND JSON_CONTAINS(params, :grantParam))").
		Param("grantParam", string(maps.Map{"grantId": grantId}.AsJSON())).
		Where("clusterId IN (SELECT id FROM edgeNodeClusters WHERE state=1)").
		Count()
}

// FindAllEnabledNodesWithGrantId 查找使用某个认证的所有节点
func (this *NodeDAO) FindAllEnabledNodesWithGrantId(tx *dbs.Tx, grantId int64) (result []*Node, err error) {
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Where("id IN (SELECT nodeId FROM edgeNodeLogins WHERE type='ssh' AND JSON_CONTAINS(params, :grantParam))").
		Param("grantParam", string(maps.Map{"grantId": grantId}.AsJSON())).
		Where("clusterId IN (SELECT id FROM edgeNodeClusters WHERE state=1)").
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// CountAllNotInstalledNodesWithClusterId 计算未安装的节点数量
func (this *NodeDAO) CountAllNotInstalledNodesWithClusterId(tx *dbs.Tx, clusterId int64) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		Attr("isInstalled", false).
		Count()
}

// FindAllNotInstalledNodesWithClusterId 查找所有未安装的节点
func (this *NodeDAO) FindAllNotInstalledNodesWithClusterId(tx *dbs.Tx, clusterId int64) (result []*Node, err error) {
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		Attr("isInstalled", false).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllLowerVersionNodesWithClusterId 计算单个集群中所有低于某个版本的节点数量
func (this *NodeDAO) CountAllLowerVersionNodesWithClusterId(tx *dbs.Tx, clusterId int64, os string, arch string, version string) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
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

// FindAllLowerVersionNodesWithClusterId 查找单个集群中所有低于某个版本的节点
func (this *NodeDAO) FindAllLowerVersionNodesWithClusterId(tx *dbs.Tx, clusterId int64, os string, arch string, version string) (result []*Node, err error) {
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		Where("status IS NOT NULL").
		Where("JSON_EXTRACT(status, '$.os')=:os").
		Where("JSON_EXTRACT(status, '$.arch')=:arch").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("os", os).
		Param("arch", arch).
		Param("version", utils.VersionToLong(version)).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllLowerVersionNodes 计算所有集群中低于某个版本的节点数量
func (this *NodeDAO) CountAllLowerVersionNodes(tx *dbs.Tx, version string) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Where("clusterId IN (SELECT id FROM "+SharedNodeClusterDAO.Table+" WHERE state=1)").
		Where("status IS NOT NULL").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("version", utils.VersionToLong(version)).
		Count()
}

// CountAllEnabledNodesWithGroupId 查找某个节点分组下的所有节点数量
func (this *NodeDAO) CountAllEnabledNodesWithGroupId(tx *dbs.Tx, groupId int64) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("groupId", groupId).
		Where("clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)").
		Count()
}

// CountAllEnabledNodesWithRegionId 查找某个节点区域下的所有节点数量
func (this *NodeDAO) CountAllEnabledNodesWithRegionId(tx *dbs.Tx, regionId int64) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("regionId", regionId).
		Where("clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)").
		Count()
}

// FindAllEnabledNodesDNSWithClusterId 获取一个集群的节点DNS信息
func (this *NodeDAO) FindAllEnabledNodesDNSWithClusterId(tx *dbs.Tx, clusterId int64, includeSecondaryNodes bool, includingLnNodes bool) (result []*Node, err error) {
	if clusterId <= 0 {
		return nil, nil
	}
	var query = this.Query(tx)
	if includeSecondaryNodes {
		query.Where("(clusterId=:primaryClusterId OR JSON_CONTAINS(secondaryClusterIds, :primaryClusterIdString))").
			Param("primaryClusterId", clusterId).
			Param("primaryClusterIdString", types.String(clusterId))
	} else {
		query.Attr("clusterId", clusterId)
	}
	if !includingLnNodes {
		query.Lte("level", 1)
	}
	_, err = query.
		State(NodeStateEnabled).
		Attr("isOn", true).
		Attr("isUp", true).
		Result("id", "name", "dnsRoutes", "isOn").
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledNodesDNSWithClusterId 计算一个集群的节点DNS数量
func (this *NodeDAO) CountAllEnabledNodesDNSWithClusterId(tx *dbs.Tx, clusterId int64) (result int64, err error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		Attr("isOn", true).
		Attr("isUp", true).
		Result("id", "name", "dnsRoutes", "isOn").
		DescPk().
		Slice(&result).
		Count()
}

// FindEnabledNodeDNS 获取单个节点的DNS信息
func (this *NodeDAO) FindEnabledNodeDNS(tx *dbs.Tx, nodeId int64) (*Node, error) {
	one, err := this.Query(tx).
		State(NodeStateEnabled).
		Pk(nodeId).
		Result("id", "name", "dnsRoutes", "clusterId", "isOn").
		Find()
	if one == nil {
		return nil, err
	}
	return one.(*Node), nil
}

// FindStatelessNodeDNS 获取单个节点的DNS信息，无论什么状态
func (this *NodeDAO) FindStatelessNodeDNS(tx *dbs.Tx, nodeId int64) (*Node, error) {
	one, err := this.Query(tx).
		Pk(nodeId).
		Result("id", "name", "dnsRoutes", "clusterId", "isOn", "state").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*Node), nil
}

// UpdateNodeDNS 修改节点的DNS信息
func (this *NodeDAO) UpdateNodeDNS(tx *dbs.Tx, nodeId int64, routes map[int64][]string) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	if routes == nil {
		routes = map[int64][]string{}
	}
	routesJSON, err := json.Marshal(routes)
	if err != nil {
		return err
	}
	var op = NewNodeOperator()
	op.Id = nodeId
	op.DnsRoutes = routesJSON
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, nodeId)
	if err != nil {
		return err
	}

	err = this.NotifyDNSUpdate(tx, nodeId)
	if err != nil {
		return err
	}

	return nil
}

// FindNodeDNSResolver 查找域名DNS Resolver
func (this *NodeDAO) FindNodeDNSResolver(tx *dbs.Tx, nodeId int64) (*nodeconfigs.DNSResolverConfig, error) {
	configJSON, err := this.Query(tx).
		Pk(nodeId).
		Result("dnsResolver").
		FindJSONCol()
	if err != nil {
		return nil, err
	}
	if IsNull(configJSON) {
		return nodeconfigs.DefaultDNSResolverConfig(), nil
	}

	var config = nodeconfigs.DefaultDNSResolverConfig()
	err = json.Unmarshal(configJSON, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// UpdateNodeDNSResolver 修改域名DNS Resolver
func (this *NodeDAO) UpdateNodeDNSResolver(tx *dbs.Tx, nodeId int64, dnsResolverConfig *nodeconfigs.DNSResolverConfig) error {
	if nodeId <= 0 {
		return ErrNotFound
	}

	configJSON, err := json.Marshal(dnsResolverConfig)
	if err != nil {
		return err
	}

	var op = NewNodeOperator()
	op.Id = nodeId
	op.DnsResolver = configJSON
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, nodeId)
}

// UpdateNodeSystem 设置系统信息
func (this *NodeDAO) UpdateNodeSystem(tx *dbs.Tx, nodeId int64, maxCPU int32) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	var op = NewNodeOperator()
	op.Id = nodeId
	op.MaxCPU = maxCPU
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, nodeId)
}

// UpdateNodeCache 设置缓存相关
func (this *NodeDAO) UpdateNodeCache(tx *dbs.Tx, nodeId int64, maxCacheDiskCapacityJSON []byte, maxCacheMemoryCapacityJSON []byte, cacheDiskDir string) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	var op = NewNodeOperator()
	op.Id = nodeId
	if len(maxCacheDiskCapacityJSON) > 0 {
		op.MaxCacheDiskCapacity = maxCacheDiskCapacityJSON
	}
	if len(maxCacheMemoryCapacityJSON) > 0 {
		op.MaxCacheMemoryCapacity = maxCacheMemoryCapacityJSON
	}
	op.CacheDiskDir = cacheDiskDir
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, nodeId)
}

// UpdateNodeUpCount 计算节点上线|下线状态
func (this *NodeDAO) UpdateNodeUpCount(tx *dbs.Tx, nodeId int64, isUp bool, maxUp int, maxDown int) (changed bool, err error) {
	if nodeId <= 0 {
		return false, errors.New("invalid nodeId")
	}
	one, err := this.Query(tx).
		Pk(nodeId).
		Result("isUp", "countUp", "countDown").
		Find()
	if err != nil {
		return false, err
	}
	if one == nil {
		return false, nil
	}
	oldIsUp := one.(*Node).IsUp

	// 如果新老状态一致，则不做任何事情
	if oldIsUp == isUp {
		return false, nil
	}

	countUp := int(one.(*Node).CountUp)
	countDown := int(one.(*Node).CountDown)

	var op = NewNodeOperator()
	op.Id = nodeId

	if isUp {
		countUp++
		countDown = 0

		if countUp >= maxUp {
			changed = true
			op.IsUp = true
		}
	} else {
		countDown++
		countUp = 0

		if countDown >= maxDown {
			changed = true
			op.IsUp = false
		}
	}

	op.CountUp = countUp
	op.CountDown = countDown
	err = this.Save(tx, op)
	if err != nil {
		return false, err
	}

	if changed {
		err = this.NotifyDNSUpdate(tx, nodeId)
		if err != nil {
			return true, err
		}
	}

	return
}

// UpdateNodeUp 设置节点上下线状态
func (this *NodeDAO) UpdateNodeUp(tx *dbs.Tx, nodeId int64, isUp bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	var op = NewNodeOperator()
	op.Id = nodeId
	op.IsUp = isUp
	op.CountUp = 0
	op.CountDown = 0
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	// TODO 只有前后状态不一致的时候才需要更新DNS

	return this.NotifyDNSUpdate(tx, nodeId)
}

// UpdateNodeActive 修改节点活跃状态
func (this *NodeDAO) UpdateNodeActive(tx *dbs.Tx, nodeId int64, isActive bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("isActive", isActive).
		Set("inactiveNotifiedAt", 0).
		Update()
	return err
}

// UpdateNodeInactiveNotifiedAt 修改节点的离线通知时间
func (this *NodeDAO) UpdateNodeInactiveNotifiedAt(tx *dbs.Tx, nodeId int64, inactiveAt int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("inactiveNotifiedAt", inactiveAt).
		Update()
	return err
}

// FindNodeInactiveNotifiedAt 读取上次的节点离线通知时间
func (this *NodeDAO) FindNodeInactiveNotifiedAt(tx *dbs.Tx, nodeId int64) (int64, error) {
	return this.Query(tx).
		Pk(nodeId).
		Result("inactiveNotifiedAt").
		FindInt64Col(0)
}

// FindNodeActive 检查节点活跃状态
func (this *NodeDAO) FindNodeActive(tx *dbs.Tx, nodeId int64) (bool, error) {
	isActive, err := this.Query(tx).
		Pk(nodeId).
		Result("isActive").
		FindIntCol(0)
	if err != nil {
		return false, err
	}
	return isActive == 1, nil
}

// FindNodeVersion 查找节点的版本号
func (this *NodeDAO) FindNodeVersion(tx *dbs.Tx, nodeId int64) (int64, error) {
	return this.Query(tx).
		Pk(nodeId).
		Result("version").
		FindInt64Col(0)
}

// GenUniqueId 生成唯一ID
func (this *NodeDAO) GenUniqueId(tx *dbs.Tx) (string, error) {
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

// FindEnabledNodesWithIds 根据一组ID查找一组节点
func (this *NodeDAO) FindEnabledNodesWithIds(tx *dbs.Tx, nodeIds []int64) (result []*Node, err error) {
	if len(nodeIds) == 0 {
		return nil, nil
	}
	idStrings := []string{}
	for _, nodeId := range nodeIds {
		idStrings = append(idStrings, numberutils.FormatInt64(nodeId))
	}
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Where("id IN ("+strings.Join(idStrings, ", ")+")").
		Result("id", "connectedAPINodes", "isActive", "isOn").
		Slice(&result).
		Reuse(false).
		FindAll()
	return
}

// DeleteNodeFromCluster 从集群中删除节点
func (this *NodeDAO) DeleteNodeFromCluster(tx *dbs.Tx, nodeId int64, clusterId int64) error {
	one, err := this.Query(tx).
		Pk(nodeId).
		Result("clusterId", "secondaryClusterIds").
		Find()
	if err != nil {
		return err
	}
	if one == nil {
		return nil
	}

	var node = one.(*Node)

	var secondaryClusterIds = []int64{}
	for _, secondaryClusterId := range node.DecodeSecondaryClusterIds() {
		if secondaryClusterId == clusterId {
			continue
		}
		secondaryClusterIds = append(secondaryClusterIds, secondaryClusterId)
	}

	var newClusterId = int64(node.ClusterId)

	if newClusterId == clusterId {
		newClusterId = 0

		// 选择一个从集群作为主集群
		if len(secondaryClusterIds) > 0 {
			newClusterId = secondaryClusterIds[0]
			secondaryClusterIds = secondaryClusterIds[1:]
		}
	}

	secondaryClusterIdsJSON, err := json.Marshal(secondaryClusterIds)
	if err != nil {
		return err
	}
	var op = NewNodeOperator()
	op.Id = nodeId
	op.ClusterId = newClusterId
	op.SecondaryClusterIds = secondaryClusterIdsJSON

	if newClusterId == 0 {
		op.State = NodeStateDisabled
	}

	return this.Save(tx, op)
}

// TransferPrimaryClusterNodes 自动转移集群下的节点
func (this *NodeDAO) TransferPrimaryClusterNodes(tx *dbs.Tx, primaryClusterId int64) error {
	if primaryClusterId <= 0 {
		return nil
	}
	ones, err := this.Query(tx).
		Attr("clusterId", primaryClusterId).
		Result("id", "secondaryClusterIds").
		State(NodeStateEnabled).
		FindAll()
	if err != nil {
		return err
	}
	for _, one := range ones {
		var node = one.(*Node)
		clusterIds := node.DecodeSecondaryClusterIds()
		if len(clusterIds) == 0 {
			continue
		}
		var clusterId = clusterIds[0]
		var secondaryClusterIds = clusterIds[1:]
		secondaryClusterIdsJSON, err := json.Marshal(secondaryClusterIds)
		if err != nil {
			return err
		}
		err = this.Query(tx).
			Pk(node.Id).
			Set("clusterId", clusterId).
			Set("secondaryClusterIds", secondaryClusterIdsJSON).
			UpdateQuickly()
		if err != nil {
			return err
		}
	}
	return nil
}

// FindParentNodeConfigs 查找父级节点配置
func (this *NodeDAO) FindParentNodeConfigs(tx *dbs.Tx, nodeId int64, groupId int64, clusterIds []int64, level int) (result map[int64][]*nodeconfigs.ParentNodeConfig, err error) {
	result = map[int64][]*nodeconfigs.ParentNodeConfig{} // clusterId => []*ParentNodeConfig

	// 当前分组的L2
	var parentNodes []*Node
	if groupId > 0 {
		parentNodes, err = this.FindEnabledNodesWithGroupIdAndLevel(tx, groupId, level+1)
		if err != nil {
			return nil, err
		}
	} else if nodeId > 0 {
		// 当前节点所属分组
		groupId, err = this.Query(tx).Result("groupId").FindInt64Col(0)
		if err != nil {
			return nil, err
		}

		if groupId > 0 {
			parentNodes, err = this.FindEnabledNodesWithGroupIdAndLevel(tx, groupId, level+1)
			if err != nil {
				return nil, err
			}
		}
	}

	// 当前集群的L2
	if len(parentNodes) == 0 {
		for _, clusterId := range clusterIds {
			clusterParentNodes, err := this.FindEnabledNodesWithClusterIdAndLevel(tx, clusterId, level+1)
			if err != nil {
				return nil, err
			}
			if len(clusterParentNodes) > 0 {
				parentNodes = append(parentNodes, clusterParentNodes...)
			}
		}
	}

	if len(parentNodes) > 0 {
		for _, node := range parentNodes {
			// 是否有Ln地址
			var addrStrings = node.DecodeLnAddrs()
			if len(addrStrings) == 0 {
				// 如果没有就取节点的可访问地址
				addrs, err := SharedNodeIPAddressDAO.FindNodeAccessAndUpIPAddresses(tx, int64(node.Id), nodeconfigs.NodeRoleNode)
				if err != nil {
					return nil, err
				}
				for _, addr := range addrs {
					if addr.IsOn {
						addrStrings = append(addrStrings, addr.DNSIP())
					}
				}
			}

			// 没有地址就跳过
			if len(addrStrings) == 0 {
				continue
			}

			var secretHash = fmt.Sprintf("%x", sha256.Sum256([]byte(node.UniqueId+"@"+node.Secret)))

			for _, clusterId := range node.AllClusterIds() {
				parentNodeConfigs, _ := result[clusterId]
				parentNodeConfigs = append(parentNodeConfigs, &nodeconfigs.ParentNodeConfig{
					Id:         int64(node.Id),
					Addrs:      addrStrings,
					SecretHash: secretHash,
				})
				result[clusterId] = parentNodeConfigs
			}
		}
	}

	return
}

// FindNodeDDoSProtection 获取节点的DDOS设置
func (this *NodeDAO) FindNodeDDoSProtection(tx *dbs.Tx, nodeId int64) (*ddosconfigs.ProtectionConfig, error) {
	one, err := this.Query(tx).
		Result("ddosProtection").
		Pk(nodeId).
		Find()
	if one == nil || err != nil {
		return nil, err
	}

	return one.(*Node).DecodeDDoSProtection(), nil
}

// UpdateNodeDDoSProtection 设置集群的DDoS设置
func (this *NodeDAO) UpdateNodeDDoSProtection(tx *dbs.Tx, nodeId int64, ddosProtection *ddosconfigs.ProtectionConfig) error {
	if nodeId <= 0 {
		return ErrNotFound
	}

	var op = NewNodeOperator()
	op.Id = nodeId

	if ddosProtection == nil {
		op.DdosProtection = "{}"
	} else {
		ddosProtectionJSON, err := json.Marshal(ddosProtection)
		if err != nil {
			return err
		}
		op.DdosProtection = ddosProtectionJSON
	}

	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	clusterId, err := this.FindNodeClusterId(tx, nodeId)
	if err != nil {
		return err
	}
	if clusterId > 0 {
		return SharedNodeTaskDAO.CreateNodeTask(tx, nodeconfigs.NodeRoleNode, clusterId, nodeId, 0, NodeTaskTypeDDosProtectionChanged, 0)
	}
	return nil
}

// NotifyUpdate 通知节点相关更新
func (this *NodeDAO) NotifyUpdate(tx *dbs.Tx, nodeId int64) error {
	// 这里只需要通知单个集群即可，因为节点是公用的，更新一个就相当于更新了所有
	clusterId, err := this.FindNodeClusterId(tx, nodeId)
	if err != nil {
		return err
	}
	err = SharedNodeTaskDAO.CreateNodeTask(tx, nodeconfigs.NodeRoleNode, clusterId, nodeId, 0, NodeTaskTypeConfigChanged, 0)
	if err != nil {
		return err
	}

	return nil
}

// NotifyLevelUpdate 通知节点级别更新
func (this *NodeDAO) NotifyLevelUpdate(tx *dbs.Tx, nodeId int64) error {
	// 这里只需要通知单个集群即可，因为节点是公用的，更新一个就相当于更新了所有
	clusterIds, err := this.FindEnabledAndOnNodeClusterIds(tx, nodeId)
	if err != nil {
		return err
	}
	for _, clusterId := range clusterIds {
		err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, 0, NodeTaskTypeNodeLevelChanged)
		if err != nil {
			return err
		}
	}
	return nil
}

// NotifyDNSUpdate 通知节点相关DNS更新
func (this *NodeDAO) NotifyDNSUpdate(tx *dbs.Tx, nodeId int64) error {
	clusterIds, err := this.FindEnabledAndOnNodeClusterIds(tx, nodeId)
	if err != nil {
		return err
	}
	for _, clusterId := range clusterIds {
		dnsInfo, err := SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId, nil)
		if err != nil {
			return err
		}
		if dnsInfo == nil {
			continue
		}
		if len(dnsInfo.DnsName) == 0 || dnsInfo.DnsDomainId <= 0 {
			continue
		}
		err = dns.SharedDNSTaskDAO.CreateNodeTask(tx, nodeId, dns.DNSTaskTypeNodeChange)
		if err != nil {
			return err
		}
	}
	return nil
}
