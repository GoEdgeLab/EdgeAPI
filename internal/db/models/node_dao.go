package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"strconv"
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

var SharedNodeDAO = NewNodeDAO()

// 启用条目
func (this *NodeDAO) EnableNode(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeDAO) DisableNode(id int64) (err error) {
	_, err = this.Query().
		Pk(id).
		Set("state", NodeStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *NodeDAO) FindEnabledNode(id int64) (*Node, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", NodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Node), err
}

// 根据主键查找名称
func (this *NodeDAO) FindNodeName(id uint32) (string, error) {
	name, err := this.Query().
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}

// 创建节点
func (this *NodeDAO) CreateNode(name string, clusterId int64) (nodeId int64, err error) {
	uniqueId, err := this.genUniqueId()
	if err != nil {
		return 0, err
	}

	secret := rands.String(32)

	// 保存API Token
	err = SharedApiTokenDAO.CreateAPIToken(uniqueId, secret, NodeRoleNode)
	if err != nil {
		return
	}

	op := NewNodeOperator()
	op.Name = name
	op.UniqueId = uniqueId
	op.Secret = secret
	op.ClusterId = clusterId
	op.IsOn = 1
	op.State = NodeStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// 修改节点
func (this *NodeDAO) UpdateNode(nodeId int64, name string, clusterId int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	op := NewNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.ClusterId = clusterId
	op.LatestVersion = dbs.SQL("latestVersion+1")
	_, err := this.Save(op)
	return err
}

// 更新节点版本
func (this *NodeDAO) UpdateNodeLatestVersion(nodeId int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	op := NewNodeOperator()
	op.Id = nodeId
	op.LatestVersion = dbs.SQL("latestVersion+1")
	_, err := this.Save(op)
	return err
}

// 批量更新节点版本
func (this *NodeDAO) UpdateAllNodesLatestVersionMatch(clusterId int64) error {
	nodeIds, err := this.FindAllNodeIdsMatch(clusterId)
	if err != nil {
		return err
	}
	if len(nodeIds) == 0 {
		return nil
	}
	_, err = this.Query().
		Pk(nodeIds).
		Set("latestVersion", dbs.SQL("latestVersion+1")).
		Update()
	return err
}

// 同步集群中的节点版本
func (this *NodeDAO) SyncNodeVersionsWithCluster(clusterId int64) error {
	if clusterId <= 0 {
		return errors.New("invalid cluster")
	}
	_, err := this.Query().
		Attr("clusterId", clusterId).
		Set("version", dbs.SQL("latestVersion")).
		Update()
	return err
}

// 取得有变更的集群
func (this *NodeDAO) FindChangedClusterIds() ([]int64, error) {
	ones, _, err := this.Query().
		State(NodeStateEnabled).
		Gt("latestVersion", 0).
		Where("version!=latestVersion").
		Result("DISTINCT(clusterId) AS clusterId").
		FindOnes()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, one.GetInt64("clusterId"))
	}
	return result, nil
}

// 计算所有节点数量
func (this *NodeDAO) CountAllEnabledNodes() (int64, error) {
	return this.Query().
		State(NodeStateEnabled).
		Count()
}

// 列出单页节点
func (this *NodeDAO) ListEnabledNodesMatch(offset int64, size int64, clusterId int64, installState configutils.BoolState, activeState configutils.BoolState) (result []*Node, err error) {
	query := this.Query().
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
		query.Where("JSON_EXTRACT(status, '$.isActive') AND JSON_EXTRACT(status, '$.updatedAt')-UNIX_TIMESTAMP()<=60")
	case configutils.BoolStateNo:
		query.Where("(status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR JSON_EXTRACT(status, '$.updatedAt')-UNIX_TIMESTAMP()>60)")
	}

	_, err = query.FindAll()
	return
}

// 根据节点ID和密钥查询节点
func (this *NodeDAO) FindEnabledNodeWithUniqueIdAndSecret(uniqueId string, secret string) (*Node, error) {
	one, err := this.Query().
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
func (this *NodeDAO) FindEnabledNodeWithUniqueId(uniqueId string) (*Node, error) {
	one, err := this.Query().
		Attr("uniqueId", uniqueId).
		State(NodeStateEnabled).
		Find()

	if one != nil {
		return one.(*Node), err
	}

	return nil, err
}

// 获取节点集群ID
func (this *NodeDAO) FindNodeClusterId(nodeId int64) (int64, error) {
	col, err := this.Query().
		Pk(nodeId).
		Result("clusterId").
		FindCol(0)
	return types.Int64(col), err
}

// 匹配节点并返回节点ID
func (this *NodeDAO) FindAllNodeIdsMatch(clusterId int64) (result []int64, err error) {
	query := this.Query()
	query.State(NodeStateEnabled)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
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

// 计算节点数量
func (this *NodeDAO) CountAllEnabledNodesMatch(clusterId int64, installState configutils.BoolState, activeState configutils.BoolState) (int64, error) {
	query := this.Query()
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
		query.Where("JSON_EXTRACT(status, '$.isActive') AND JSON_EXTRACT(status, '$.updatedAt')-UNIX_TIMESTAMP()<=60")
	case configutils.BoolStateNo:
		query.Where("(status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR JSON_EXTRACT(status, '$.updatedAt')-UNIX_TIMESTAMP()>60)")
	}

	return query.Count()
}

// 更改节点状态
func (this *NodeDAO) UpdateNodeStatus(nodeId int64, statusJSON []byte) error {
	_, err := this.Query().
		Pk(nodeId).
		Set("status", string(statusJSON)).
		Update()
	return err
}

// 设置节点安装状态
func (this *NodeDAO) UpdateNodeIsInstalled(nodeId int64, isInstalled bool) error {
	_, err := this.Query().
		Pk(nodeId).
		Set("isInstalled", isInstalled).
		Set("installStatus", "null"). // 重置安装状态
		Update()
	return err
}

// 查询节点的安装状态
func (this *NodeDAO) FindNodeInstallStatus(nodeId int64) (*NodeInstallStatus, error) {
	installStatus, err := this.Query().
		Pk(nodeId).
		Result("installStatus").
		FindStringCol("")
	if err != nil {
		return nil, err
	}

	if len(installStatus) == 0 {
		return NewNodeInstallStatus(), nil
	}

	status := &NodeInstallStatus{}
	err = json.Unmarshal([]byte(installStatus), status)
	return status, err
}

// 修改节点的安装状态
func (this *NodeDAO) UpdateNodeInstallStatus(nodeId int64, status *NodeInstallStatus) error {
	if status == nil {
		_, err := this.Query().
			Pk(nodeId).
			Set("installStatus", "null").
			Update()
		return err
	}

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	_, err = this.Query().
		Pk(nodeId).
		Set("installStatus", string(data)).
		Update()
	return err
}

// 组合配置
func (this *NodeDAO) ComposeNodeConfig(nodeId int64) (*nodeconfigs.NodeConfig, error) {
	node, err := this.FindEnabledNode(nodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found '" + strconv.FormatInt(nodeId, 10) + "'")
	}

	config := &nodeconfigs.NodeConfig{
		Id:      node.UniqueId,
		IsOn:    node.IsOn == 1,
		Servers: nil,
		Version: int64(node.Version),
		Name:    node.Name,
	}

	// 获取所有的服务
	servers, err := SharedServerDAO.FindAllEnabledServersWithNode(int64(node.Id))
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
	settingJSON, err := SharedSysSettingDAO.ReadSetting(SettingCodeGlobalConfig)
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

	return config, nil
}

// 生成唯一ID
func (this *NodeDAO) genUniqueId() (string, error) {
	for {
		uniqueId := rands.HexString(32)
		ok, err := this.Query().
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
