package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

const (
	ServerGroupStateEnabled  = 1 // 已启用
	ServerGroupStateDisabled = 0 // 已禁用
)

type ServerGroupDAO dbs.DAO

func NewServerGroupDAO() *ServerGroupDAO {
	return dbs.NewDAO(&ServerGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerGroups",
			Model:  new(ServerGroup),
			PkName: "id",
		},
	}).(*ServerGroupDAO)
}

var SharedServerGroupDAO *ServerGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedServerGroupDAO = NewServerGroupDAO()
	})
}

// EnableServerGroup 启用条目
func (this *ServerGroupDAO) EnableServerGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ServerGroupStateEnabled).
		Update()
	return err
}

// DisableServerGroup 禁用条目
func (this *ServerGroupDAO) DisableServerGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ServerGroupStateDisabled).
		Update()
	return err
}

// FindEnabledServerGroup 查找启用中的条目
func (this *ServerGroupDAO) FindEnabledServerGroup(tx *dbs.Tx, id int64) (*ServerGroup, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ServerGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ServerGroup), err
}

// FindServerGroupName 根据主键查找名称
func (this *ServerGroupDAO) FindServerGroupName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateGroup 创建分组
func (this *ServerGroupDAO) CreateGroup(tx *dbs.Tx, name string, userId int64) (groupId int64, err error) {
	var op = NewServerGroupOperator()
	op.State = ServerGroupStateEnabled
	op.Name = name
	op.UserId = userId
	op.IsOn = true
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateGroup 修改分组
func (this *ServerGroupDAO) UpdateGroup(tx *dbs.Tx, groupId int64, name string) error {
	if groupId <= 0 {
		return errors.New("invalid groupId")
	}
	var op = NewServerGroupOperator()
	op.Id = groupId
	op.Name = name
	err := this.Save(tx, op)
	return err
}

// FindAllEnabledGroups 查找所有分组
func (this *ServerGroupDAO) FindAllEnabledGroups(tx *dbs.Tx, userId int64) (result []*ServerGroup, err error) {
	var query = this.Query(tx).
		State(ServerGroupStateEnabled)
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		query.Attr("userId", 0)
	}
	_, err = query.
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// UpdateGroupOrders 修改分组排序
func (this *ServerGroupDAO) UpdateGroupOrders(tx *dbs.Tx, groupIds []int64, userId int64) error {
	for index, groupId := range groupIds {
		var query = this.Query(tx)
		if userId > 0 {
			query.Attr("userId", userId)
		}
		_, err := query.
			Pk(groupId).
			Set("order", len(groupIds)-index).
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}

// FindHTTPReverseProxyRef 根据条件获取HTTP反向代理配置
func (this *ServerGroupDAO) FindHTTPReverseProxyRef(tx *dbs.Tx, groupId int64) (*serverconfigs.ReverseProxyRef, error) {
	reverseProxy, err := this.Query(tx).
		Pk(groupId).
		Result("httpReverseProxy").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(reverseProxy) == 0 || reverseProxy == "null" {
		return nil, nil
	}
	config := &serverconfigs.ReverseProxyRef{}
	err = json.Unmarshal([]byte(reverseProxy), config)
	return config, err
}

// FindTCPReverseProxyRef 根据条件获取TCP反向代理配置
func (this *ServerGroupDAO) FindTCPReverseProxyRef(tx *dbs.Tx, groupId int64) (*serverconfigs.ReverseProxyRef, error) {
	reverseProxy, err := this.Query(tx).
		Pk(groupId).
		Result("tcpReverseProxy").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(reverseProxy) == 0 || reverseProxy == "null" {
		return nil, nil
	}
	config := &serverconfigs.ReverseProxyRef{}
	err = json.Unmarshal([]byte(reverseProxy), config)
	return config, err
}

// FindUDPReverseProxyRef 根据条件获取UDP反向代理配置
func (this *ServerGroupDAO) FindUDPReverseProxyRef(tx *dbs.Tx, groupId int64) (*serverconfigs.ReverseProxyRef, error) {
	reverseProxy, err := this.Query(tx).
		Pk(groupId).
		Result("udpReverseProxy").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(reverseProxy) == 0 || reverseProxy == "null" {
		return nil, nil
	}
	config := &serverconfigs.ReverseProxyRef{}
	err = json.Unmarshal([]byte(reverseProxy), config)
	return config, err
}

// UpdateHTTPReverseProxy 修改HTTP反向代理配置
func (this *ServerGroupDAO) UpdateHTTPReverseProxy(tx *dbs.Tx, groupId int64, config []byte) error {
	if groupId <= 0 {
		return errors.New("groupId should not be smaller than 0")
	}
	var op = NewServerGroupOperator()
	op.Id = groupId
	op.HttpReverseProxy = JSONBytes(config)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, groupId)
}

// UpdateTCPReverseProxy 修改TCP反向代理配置
func (this *ServerGroupDAO) UpdateTCPReverseProxy(tx *dbs.Tx, groupId int64, config []byte) error {
	if groupId <= 0 {
		return errors.New("groupId should not be smaller than 0")
	}
	var op = NewServerGroupOperator()
	op.Id = groupId
	op.TcpReverseProxy = JSONBytes(config)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, groupId)
}

// UpdateUDPReverseProxy 修改UDP反向代理配置
func (this *ServerGroupDAO) UpdateUDPReverseProxy(tx *dbs.Tx, groupId int64, config []byte) error {
	if groupId <= 0 {
		return errors.New("groupId should not be smaller than 0")
	}
	var op = NewServerGroupOperator()
	op.Id = groupId
	op.UdpReverseProxy = JSONBytes(config)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, groupId)
}

// FindGroupWebId 查找分组WebId
func (this *ServerGroupDAO) FindGroupWebId(tx *dbs.Tx, groupId int64) (webId int64, err error) {
	return this.Query(tx).
		Pk(groupId).
		Result("webId").
		FindInt64Col(0)
}

// FindEnabledGroupIdWithWebId 根据WebId查找分组
func (this *ServerGroupDAO) FindEnabledGroupIdWithWebId(tx *dbs.Tx, webId int64) (int64, error) {
	if webId <= 0 {
		return 0, nil
	}
	return this.Query(tx).
		State(ServerGroupStateEnabled).
		ResultPk().
		Attr("webId", webId).
		FindInt64Col(0)
}

// InitGroupWeb 初始化Web配置
func (this *ServerGroupDAO) InitGroupWeb(tx *dbs.Tx, groupId int64) (int64, error) {
	if groupId <= 0 {
		return 0, errors.New("invalid groupId")
	}

	webId, err := SharedHTTPWebDAO.CreateWeb(tx, 0, 0, nil)
	if err != nil {
		return 0, err
	}

	err = this.Query(tx).
		Pk(groupId).
		Set("webId", webId).
		UpdateQuickly()
	if err != nil {
		return 0, err
	}

	return webId, nil
}

// ComposeGroupConfig 组合配置
func (this *ServerGroupDAO) ComposeGroupConfig(tx *dbs.Tx, groupId int64, cacheMap *utils.CacheMap) (*serverconfigs.ServerGroupConfig, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}

	var cacheKey = this.Table + ":config:" + types.String(groupId)
	var cacheConfig, _ = cacheMap.Get(cacheKey)
	if cacheConfig != nil {
		// 克隆，防止分解后的Server配置相互受到影响
		configJSON, err := json.Marshal(cacheConfig)
		if err != nil {
			return nil, err
		}

		var clonedConfig = &serverconfigs.ServerGroupConfig{}
		err = json.Unmarshal(configJSON, clonedConfig)
		if err != nil {
			return nil, err
		}
		return clonedConfig, nil
	}

	group, err := this.FindEnabledServerGroup(tx, groupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	var config = &serverconfigs.ServerGroupConfig{
		Id:   int64(group.Id),
		Name: group.Name,
		IsOn: group.IsOn,
	}

	if IsNotNull(group.HttpReverseProxy) {
		reverseProxyRef := &serverconfigs.ReverseProxyRef{}
		err := json.Unmarshal(group.HttpReverseProxy, reverseProxyRef)
		if err != nil {
			return nil, err
		}
		config.HTTPReverseProxyRef = reverseProxyRef

		reverseProxyConfig, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, cacheMap)
		if err != nil {
			return nil, err
		}
		if reverseProxyConfig != nil {
			config.HTTPReverseProxy = reverseProxyConfig
		}
	}

	if IsNotNull(group.TcpReverseProxy) {
		reverseProxyRef := &serverconfigs.ReverseProxyRef{}
		err := json.Unmarshal(group.TcpReverseProxy, reverseProxyRef)
		if err != nil {
			return nil, err
		}
		config.TCPReverseProxyRef = reverseProxyRef

		reverseProxyConfig, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, cacheMap)
		if err != nil {
			return nil, err
		}
		if reverseProxyConfig != nil {
			config.TCPReverseProxy = reverseProxyConfig
		}
	}

	if IsNotNull(group.UdpReverseProxy) {
		reverseProxyRef := &serverconfigs.ReverseProxyRef{}
		err := json.Unmarshal(group.UdpReverseProxy, reverseProxyRef)
		if err != nil {
			return nil, err
		}
		config.UDPReverseProxyRef = reverseProxyRef

		reverseProxyConfig, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, cacheMap)
		if err != nil {
			return nil, err
		}
		if reverseProxyConfig != nil {
			config.UDPReverseProxy = reverseProxyConfig
		}
	}

	// web
	if group.WebId > 0 {
		webConfig, err := SharedHTTPWebDAO.ComposeWebConfig(tx, int64(group.WebId), cacheMap)
		if err != nil {
			return nil, err
		}
		if webConfig != nil {
			config.Web = webConfig
		}
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// FindEnabledGroupIdWithReverseProxyId 查找包含某个反向代理的服务分组
func (this *ServerGroupDAO) FindEnabledGroupIdWithReverseProxyId(tx *dbs.Tx, reverseProxyId int64) (serverId int64, err error) {
	return this.Query(tx).
		State(ServerStateEnabled).
		Where("(JSON_CONTAINS(httpReverseProxy, :jsonQuery) OR JSON_CONTAINS(tcpReverseProxy, :jsonQuery) OR JSON_CONTAINS(udpReverseProxy, :jsonQuery))").
		Param("jsonQuery", maps.Map{"reverseProxyId": reverseProxyId}.AsJSON()).
		ResultPk().
		FindInt64Col(0)
}

// CheckUserGroup 检查用户分组
func (this *ServerGroupDAO) CheckUserGroup(tx *dbs.Tx, userId int64, groupId int64) error {
	b, err := this.Query(tx).
		Pk(groupId).
		Attr("userId", userId).
		State(ServerGroupStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !b {
		return ErrNotFound
	}
	return nil
}

// ExistsGroup 检查分组ID是否存在
func (this *ServerGroupDAO) ExistsGroup(tx *dbs.Tx, groupId int64) (bool, error) {
	if groupId <= 0 {
		return false, nil
	}
	return this.Query(tx).
		Pk(groupId).
		State(ServerGroupStateEnabled).
		Exist()
}

// NotifyUpdate 通知更新
func (this *ServerGroupDAO) NotifyUpdate(tx *dbs.Tx, groupId int64) error {
	serverIds, err := SharedServerDAO.FindAllEnabledServerIdsWithGroupId(tx, groupId)
	if err != nil {
		return err
	}
	for _, serverId := range serverIds {
		err = SharedServerDAO.NotifyUpdate(tx, serverId)
		if err != nil {
			return err
		}
	}
	return nil
}
