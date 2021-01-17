package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"strconv"
	"strings"
)

const (
	NodeStateEnabled  = 1 // 已启用
	NodeStateDisabled = 0 // 已禁用
)

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

// 启用条目
func (this *NodeDAO) EnableNode(tx *dbs.Tx, id uint32) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", NodeStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeDAO) DisableNode(tx *dbs.Tx, nodeId int64) (err error) {
	_, err = this.Query(tx).
		Pk(nodeId).
		Set("state", NodeStateDisabled).
		Update()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, nodeId)
}

// 查找启用中的条目
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

// 根据主键查找名称
func (this *NodeDAO) FindNodeName(tx *dbs.Tx, id int64) (string, error) {
	name, err := this.Query(tx).
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}

// 创建节点
func (this *NodeDAO) CreateNode(tx *dbs.Tx, adminId int64, name string, clusterId int64, groupId int64, regionId int64) (nodeId int64, err error) {
	uniqueId, err := this.genUniqueId(tx)
	if err != nil {
		return 0, err
	}

	secret := rands.String(32)

	// 保存API Token
	err = SharedApiTokenDAO.CreateAPIToken(tx, uniqueId, secret, NodeRoleNode)
	if err != nil {
		return
	}

	op := NewNodeOperator()
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

	return types.Int64(op.Id), nil
}

// 修改节点
func (this *NodeDAO) UpdateNode(tx *dbs.Tx, nodeId int64, name string, clusterId int64, groupId int64, regionId int64, maxCPU int32, isOn bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	op := NewNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.ClusterId = clusterId
	op.GroupId = groupId
	op.RegionId = regionId
	op.LatestVersion = dbs.SQL("latestVersion+1")
	op.MaxCPU = maxCPU
	op.IsOn = isOn
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, nodeId)
}

// 计算所有节点数量
func (this *NodeDAO) CountAllEnabledNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Count()
}

// 列出单页节点
func (this *NodeDAO) ListEnabledNodesMatch(tx *dbs.Tx, offset int64, size int64, clusterId int64, installState configutils.BoolState, activeState configutils.BoolState, keyword string, groupId int64, regionId int64) (result []*Node, err error) {
	query := this.Query(tx).
		State(NodeStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result)

	// 集群
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

	// 关键词
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR JSON_EXTRACT(status,'$.hostname') LIKE :keyword OR id IN (SELECT nodeId FROM "+SharedNodeIPAddressDAO.Table+" WHERE ip LIKE :keyword))").
			Param("keyword", "%"+keyword+"%")
	}

	// 分组
	if groupId > 0 {
		query.Attr("groupId", groupId)
	}

	// 区域
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}

	_, err = query.FindAll()
	return
}

// 根据节点ID和密钥查询节点
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

// 根据节点ID获取节点
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

// 获取节点集群ID
func (this *NodeDAO) FindNodeClusterId(tx *dbs.Tx, nodeId int64) (int64, error) {
	col, err := this.Query(tx).
		Pk(nodeId).
		Result("clusterId").
		FindCol(0)
	return types.Int64(col), err
}

// 匹配节点并返回节点ID
func (this *NodeDAO) FindAllNodeIdsMatch(tx *dbs.Tx, clusterId int64, isOn configutils.BoolState) (result []int64, err error) {
	query := this.Query(tx)
	query.State(NodeStateEnabled)
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

// 获取一个集群的所有节点
func (this *NodeDAO) FindAllEnabledNodesWithClusterId(tx *dbs.Tx, clusterId int64) (result []*Node, err error) {
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 取得一个集群离线的节点
func (this *NodeDAO) FindAllInactiveNodesWithClusterId(tx *dbs.Tx, clusterId int64) (result []*Node, err error) {
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		Attr("isOn", true). // 只监控启用的节点
		Attr("isInstalled", true). // 只监控已经安装的节点
		Attr("isActive", true). // 当前已经在线的
		Where("(status IS NULL OR (JSON_EXTRACT(status, '$.isActive')=false AND UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>10) OR  UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>120)").
		Result("id").
		Slice(&result).
		FindAll()
	return
}

// 计算节点数量
func (this *NodeDAO) CountAllEnabledNodesMatch(tx *dbs.Tx, clusterId int64, installState configutils.BoolState, activeState configutils.BoolState, keyword string, groupId int64, regionId int64) (int64, error) {
	query := this.Query(tx)
	query.State(NodeStateEnabled)

	// 集群
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

	// 关键词
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR JSON_EXTRACT(status,'$.hostname') LIKE :keyword OR id IN (SELECT nodeId FROM "+SharedNodeIPAddressDAO.Table+" WHERE ip LIKE :keyword))").
			Param("keyword", "%"+keyword+"%")
	}

	// 分组
	if groupId > 0 {
		query.Attr("groupId", groupId)
	}

	// 区域
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}

	return query.Count()
}

// 更改节点状态
func (this *NodeDAO) UpdateNodeStatus(tx *dbs.Tx, nodeId int64, statusJSON []byte) error {
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("status", string(statusJSON)).
		Update()
	return err
}

// 更改节点在线状态
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

// 设置节点安装状态
func (this *NodeDAO) UpdateNodeIsInstalled(tx *dbs.Tx, nodeId int64, isInstalled bool) error {
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("isInstalled", isInstalled).
		Set("installStatus", "null"). // 重置安装状态
		Update()
	return err
}

// 查询节点的安装状态
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
	isInstalled := node.(*Node).IsInstalled == 1
	if len(installStatus) == 0 {
		return NewNodeInstallStatus(), nil
	}

	status := &NodeInstallStatus{}
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

// 修改节点的安装状态
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

// 组合配置
// TODO 提升运行速度
func (this *NodeDAO) ComposeNodeConfig(tx *dbs.Tx, nodeId int64) (*nodeconfigs.NodeConfig, error) {
	node, err := this.FindEnabledNode(tx, nodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found '" + strconv.FormatInt(nodeId, 10) + "'")
	}

	config := &nodeconfigs.NodeConfig{
		Id:       int64(node.Id),
		NodeId:   node.UniqueId,
		IsOn:     node.IsOn == 1,
		Servers:  nil,
		Version:  int64(node.Version),
		Name:     node.Name,
		MaxCPU:   types.Int32(node.MaxCPU),
		RegionId: int64(node.RegionId),
	}

	// 获取所有的服务
	servers, err := SharedServerDAO.FindAllEnabledServersWithNode(tx, int64(node.Id))
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if len(server.Config) == 0 {
			continue
		}

		serverConfig := &serverconfigs.ServerConfig{}
		err = json.Unmarshal([]byte(server.Config), serverConfig)
		if err != nil {
			return nil, err
		}
		config.Servers = append(config.Servers, serverConfig)
	}

	// 全局设置
	// TODO 根据用户的不同读取不同的全局设置
	settingJSON, err := SharedSysSettingDAO.ReadSetting(tx, SettingCodeServerGlobalConfig)
	if err != nil {
		return nil, err
	}
	if len(settingJSON) > 0 {
		globalConfig := &serverconfigs.GlobalConfig{}
		err = json.Unmarshal(settingJSON, globalConfig)
		if err != nil {
			return nil, err
		}
		config.GlobalConfig = globalConfig
	}

	// WAF
	clusterId := int64(node.ClusterId)
	httpFirewallPolicyId, err := SharedNodeClusterDAO.FindClusterHTTPFirewallPolicyId(tx, clusterId)
	if err != nil {
		return nil, err
	}
	if httpFirewallPolicyId > 0 {
		firewallPolicy, err := SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, httpFirewallPolicyId)
		if err != nil {
			return nil, err
		}
		if firewallPolicy != nil {
			config.HTTPFirewallPolicy = firewallPolicy
		}
	}

	// 缓存策略
	httpCachePolicyId, err := SharedNodeClusterDAO.FindClusterHTTPCachePolicyId(tx, clusterId)
	if err != nil {
		return nil, err
	}
	if httpCachePolicyId > 0 {
		cachePolicy, err := SharedHTTPCachePolicyDAO.ComposeCachePolicy(tx, httpCachePolicyId)
		if err != nil {
			return nil, err
		}
		if cachePolicy != nil {
			config.HTTPCachePolicy = cachePolicy
		}
	}

	// TOA
	toaConfig, err := SharedNodeClusterDAO.FindClusterTOAConfig(tx, clusterId)
	if err != nil {
		return nil, err
	}
	config.TOA = toaConfig

	// 系统服务
	services, err := SharedNodeClusterDAO.FindNodeClusterSystemServices(tx, clusterId)
	if err != nil {
		return nil, err
	}
	if len(services) > 0 {
		config.SystemServices = services
	}

	return config, nil
}

// 修改当前连接的API节点
func (this *NodeDAO) UpdateNodeConnectedAPINodes(tx *dbs.Tx, nodeId int64, apiNodeIds []int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	op := NewNodeOperator()
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

// 根据UniqueId获取ID
// TODO 增加缓存
func (this *NodeDAO) FindEnabledNodeIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("uniqueId", uniqueId).
		ResultPk().
		FindInt64Col(0)
}

// 计算使用某个认证的节点数量
func (this *NodeDAO) CountAllEnabledNodesWithGrantId(tx *dbs.Tx, grantId int64) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Where("id IN (SELECT nodeId FROM edgeNodeLogins WHERE type='ssh' AND JSON_CONTAINS(params, :grantParam))").
		Param("grantParam", string(maps.Map{"grantId": grantId}.AsJSON())).
		Where("clusterId IN (SELECT id FROM edgeNodeClusters WHERE state=1)").
		Count()
}

// 查找使用某个认证的所有节点
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

// 查找所有未安装的节点
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

// 计算所有低于某个版本的节点数量
func (this *NodeDAO) CountAllLowerVersionNodesWithClusterId(tx *dbs.Tx, clusterId int64, os string, arch string, version string) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		Where("status IS NOT NULL").
		Where("JSON_EXTRACT(status, '$.os')=:os").
		Where("JSON_EXTRACT(status, '$.arch')=:arch").
		Where("INET_ATON(JSON_UNQUOTE(JSON_EXTRACT(status, '$.buildVersion')))<INET_ATON(:version)").
		Param("os", os).
		Param("arch", arch).
		Param("version", version).
		Count()
}

// 查找所有低于某个版本的节点
func (this *NodeDAO) FindAllLowerVersionNodesWithClusterId(tx *dbs.Tx, clusterId int64, os string, arch string, version string) (result []*Node, err error) {
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		Where("status IS NOT NULL").
		Where("JSON_EXTRACT(status, '$.os')=:os").
		Where("JSON_EXTRACT(status, '$.arch')=:arch").
		Where("INET_ATON(JSON_UNQUOTE(JSON_EXTRACT(status, '$.buildVersion')))<INET_ATON(:version)").
		Param("os", os).
		Param("arch", arch).
		Param("version", version).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 查找某个节点分组下的所有节点数量
func (this *NodeDAO) CountAllEnabledNodesWithGroupId(tx *dbs.Tx, groupId int64) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("groupId", groupId).
		Count()
}

// 查找某个节点区域下的所有节点数量
func (this *NodeDAO) CountAllEnabledNodesWithRegionId(tx *dbs.Tx, regionId int64) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Attr("regionId", regionId).
		Count()
}

// 获取一个集群的节点DNS信息
func (this *NodeDAO) FindAllEnabledNodesDNSWithClusterId(tx *dbs.Tx, clusterId int64) (result []*Node, err error) {
	_, err = this.Query(tx).
		State(NodeStateEnabled).
		Attr("clusterId", clusterId).
		Attr("isOn", true).
		Attr("isUp", true).
		Result("id", "name", "dnsRoutes", "isOn").
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 计算一个集群的节点DNS数量
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

// 获取单个节点的DNS信息
func (this *NodeDAO) FindEnabledNodeDNS(tx *dbs.Tx, nodeId int64) (*Node, error) {
	one, err := this.Query(tx).
		State(NodeStateEnabled).
		Pk(nodeId).
		Result("id", "name", "dnsRoutes", "clusterId", "isOn").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*Node), nil
}

// 修改节点的DNS信息
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
	op := NewNodeOperator()
	op.Id = nodeId
	op.DnsRoutes = routesJSON
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, nodeId)
}

// 计算节点上线|下线状态
func (this *NodeDAO) UpdateNodeUp(tx *dbs.Tx, nodeId int64, isUp bool, maxUp int, maxDown int) (changed bool, err error) {
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
	oldIsUp := one.(*Node).IsUp == 1

	// 如果新老状态一致，则不做任何事情
	if oldIsUp == isUp {
		return false, nil
	}

	countUp := int(one.(*Node).CountUp)
	countDown := int(one.(*Node).CountDown)

	op := NewNodeOperator()
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

	return
}

// 修改节点活跃状态
func (this *NodeDAO) UpdateNodeActive(tx *dbs.Tx, nodeId int64, isActive bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("isActive", isActive).
		Update()
	return err
}

// 检查节点活跃状态
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

// 查找节点的版本号
func (this *NodeDAO) FindNodeVersion(tx *dbs.Tx, nodeId int64) (int64, error) {
	return this.Query(tx).
		Pk(nodeId).
		Result("version").
		FindInt64Col(0)
}

// 生成唯一ID
func (this *NodeDAO) genUniqueId(tx *dbs.Tx) (string, error) {
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

// 根据一组ID查找一组节点
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

// 通知更新
func (this *NodeDAO) NotifyUpdate(tx *dbs.Tx, nodeId int64) error {
	clusterId, err := this.FindNodeClusterId(tx, nodeId)
	if err != nil {
		return err
	}
	if clusterId > 0 {
		return SharedNodeTaskDAO.CreateNodeTask(tx, clusterId, nodeId, NodeTaskTypeConfigChanged)
	}
	return nil
}
