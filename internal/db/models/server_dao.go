package models

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"strconv"
	"strings"
)

const (
	ServerStateEnabled  = 1 // 已启用
	ServerStateDisabled = 0 // 已禁用
)

type ServerDAO dbs.DAO

func NewServerDAO() *ServerDAO {
	return dbs.NewDAO(&ServerDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServers",
			Model:  new(Server),
			PkName: "id",
		},
	}).(*ServerDAO)
}

var SharedServerDAO *ServerDAO

func init() {
	dbs.OnReady(func() {
		SharedServerDAO = NewServerDAO()
	})
}

// 初始化
func (this *ServerDAO) Init() {
	this.DAOObject.Init()

	// 这里不处理增删改事件，是为了避免Server修改本身的时候，也要触发别的Server变更
}

// 启用条目
func (this *ServerDAO) EnableServer(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", ServerStateEnabled).
		Update()
}

// 禁用条目
func (this *ServerDAO) DisableServer(id int64) (err error) {
	_, err = this.Query().
		Pk(id).
		Set("state", ServerStateDisabled).
		Update()
	return
}

// 查找启用中的条目
func (this *ServerDAO) FindEnabledServer(id int64) (*Server, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ServerStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Server), err
}

// 查找服务类型
func (this *ServerDAO) FindEnabledServerType(serverId int64) (string, error) {
	return this.Query().
		Pk(serverId).
		Result("type").
		FindStringCol("")
}

// 创建服务
func (this *ServerDAO) CreateServer(adminId int64, userId int64, serverType serverconfigs.ServerType, name string, description string, serverNamesJSON string, httpJSON string, httpsJSON string, tcpJSON string, tlsJSON string, unixJSON string, udpJSON string, webId int64, reverseProxyJSON []byte, clusterId int64, includeNodesJSON string, excludeNodesJSON string, groupIds []int64) (serverId int64, err error) {
	op := NewServerOperator()
	op.UserId = userId
	op.AdminId = adminId
	op.Name = name
	op.Type = serverType
	op.Description = description

	if IsNotNull(serverNamesJSON) {
		op.ServerNames = serverNamesJSON
	}
	if IsNotNull(httpJSON) {
		op.Http = httpJSON
	}
	if IsNotNull(httpsJSON) {
		op.Https = httpsJSON
	}
	if IsNotNull(tcpJSON) {
		op.Tcp = tcpJSON
	}
	if IsNotNull(tlsJSON) {
		op.Tls = tlsJSON
	}
	if IsNotNull(unixJSON) {
		op.Unix = unixJSON
	}
	if IsNotNull(udpJSON) {
		op.Udp = udpJSON
	}
	op.WebId = webId
	if len(reverseProxyJSON) > 0 {
		op.ReverseProxy = reverseProxyJSON
	}

	op.ClusterId = clusterId
	if len(includeNodesJSON) > 0 {
		op.IncludeNodes = includeNodesJSON
	}
	if len(excludeNodesJSON) > 0 {
		op.ExcludeNodes = excludeNodesJSON
	}

	if len(groupIds) == 0 {
		op.GroupIds = "[]"
	} else {
		groupIdsJSON, err := json.Marshal(groupIds)
		if err != nil {
			return 0, err
		}
		op.GroupIds = groupIdsJSON
	}

	dnsName, err := this.genDNSName()
	if err != nil {
		return 0, err
	}
	op.DnsName = dnsName

	op.Version = 1
	op.IsOn = 1
	op.State = ServerStateEnabled
	_, err = this.Save(op)

	if err != nil {
		return 0, err
	}

	serverId = types.Int64(op.Id)

	_, err = this.RenewServerConfig(serverId, false)
	if err != nil {
		return serverId, err
	}

	err = this.createEvent()
	if err != nil {
		return serverId, err
	}

	return serverId, nil
}

// 修改服务基本信息
func (this *ServerDAO) UpdateServerBasic(serverId int64, name string, description string, clusterId int64, isOn bool, groupIds []int64) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	op := NewServerOperator()
	op.Id = serverId
	op.Name = name
	op.Description = description
	op.ClusterId = clusterId
	op.IsOn = isOn

	if len(groupIds) == 0 {
		op.GroupIds = "[]"
	} else {
		groupIdsJSON, err := json.Marshal(groupIds)
		if err != nil {
			return err
		}
		op.GroupIds = groupIdsJSON
	}

	_, err := this.Save(op)
	if err != nil {
		return err
	}

	_, err = this.RenewServerConfig(serverId, false)
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 修改服务配置
func (this *ServerDAO) UpdateServerConfig(serverId int64, configJSON []byte, updateMd5 bool) (isChanged bool, err error) {
	if serverId <= 0 {
		return false, errors.New("serverId should not be smaller than 0")
	}

	// 查询以前的md5
	oldConfigMd5, err := this.Query().
		Pk(serverId).
		Result("configMd5").
		FindStringCol("")
	if err != nil {
		return false, err
	}

	globalConfig, err := SharedSysSettingDAO.ReadSetting(SettingCodeServerGlobalConfig)
	if err != nil {
		return false, err
	}

	m := md5.New()
	_, _ = m.Write(configJSON)   // 当前服务配置
	_, _ = m.Write(globalConfig) // 全局配置
	h := m.Sum(nil)
	newConfigMd5 := fmt.Sprintf("%x", h)

	// 如果配置相同则不更改
	if oldConfigMd5 == newConfigMd5 {
		return false, nil
	}

	op := NewServerOperator()
	op.Id = serverId
	op.Config = JSONBytes(configJSON)
	op.Version = dbs.SQL("version+1")

	if updateMd5 {
		op.ConfigMd5 = newConfigMd5
	}
	_, err = this.Save(op)
	return true, err
}

// 修改HTTP配置
func (this *ServerDAO) UpdateServerHTTP(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("http", string(config)).
		Update()
	if err != nil {
		return err
	}

	_, err = this.RenewServerConfig(serverId, false)
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 修改HTTPS配置
func (this *ServerDAO) UpdateServerHTTPS(serverId int64, httpsJSON []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(httpsJSON) == 0 {
		httpsJSON = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("https", string(httpsJSON)).
		Update()
	if err != nil {
		return err
	}

	_, err = this.RenewServerConfig(serverId, false)
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 修改TCP配置
func (this *ServerDAO) UpdateServerTCP(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("tcp", string(config)).
		Update()
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 修改TLS配置
func (this *ServerDAO) UpdateServerTLS(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("tls", string(config)).
		Update()
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 修改Unix配置
func (this *ServerDAO) UpdateServerUnix(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("unix", string(config)).
		Update()
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 修改UDP配置
func (this *ServerDAO) UpdateServerUDP(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("udp", string(config)).
		Update()
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 修改Web配置
func (this *ServerDAO) UpdateServerWeb(serverId int64, webId int64) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("webId", webId).
		Update()
	if err != nil {
		return err
	}
	return this.createEvent()
}

// 初始化Web配置
func (this *ServerDAO) InitServerWeb(serverId int64) (int64, error) {
	if serverId <= 0 {
		return 0, errors.New("serverId should not be smaller than 0")
	}

	webId, err := SharedHTTPWebDAO.CreateWeb(nil)
	if err != nil {
		return 0, err
	}

	_, err = this.Query().
		Pk(serverId).
		Set("webId", webId).
		Update()
	if err != nil {
		return 0, err
	}

	err = this.createEvent()
	if err != nil {
		return webId, err
	}

	return webId, nil
}

// 修改ServerNames配置
func (this *ServerDAO) UpdateServerNames(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("serverNames", string(config)).
		Update()
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 修改反向代理配置
func (this *ServerDAO) UpdateServerReverseProxy(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	op := NewServerOperator()
	op.Id = serverId
	op.ReverseProxy = JSONBytes(config)
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.createEvent()
}

// 计算所有可用服务数量
func (this *ServerDAO) CountAllEnabledServersMatch(groupId int64, keyword string) (int64, error) {
	query := this.Query().
		State(ServerStateEnabled)
	if groupId > 0 {
		query.Where("JSON_CONTAINS(groupIds, :groupId)").
			Param("groupId", numberutils.FormatInt64(groupId))
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR serverNames LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	return query.Count()
}

// 列出单页的服务
func (this *ServerDAO) ListEnabledServersMatch(offset int64, size int64, groupId int64, keyword string) (result []*Server, err error) {
	query := this.Query().
		State(ServerStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result)

	if groupId > 0 {
		query.Where("JSON_CONTAINS(groupIds, :groupId)").
			Param("groupId", numberutils.FormatInt64(groupId))
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR serverNames LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}

	_, err = query.FindAll()
	return
}

// 获取节点中的所有服务
func (this *ServerDAO) FindAllEnabledServersWithNode(nodeId int64) (result []*Server, err error) {
	// 节点所在集群
	clusterId, err := SharedNodeDAO.FindNodeClusterId(nodeId)
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, nil
	}

	_, err = this.Query().
		Attr("clusterId", clusterId).
		State(ServerStateEnabled).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 获取所有的服务ID
func (this *ServerDAO) FindAllEnabledServerIds() (serverIds []int64, err error) {
	ones, err := this.Query().
		State(ServerStateEnabled).
		AscPk().
		ResultPk().
		FindAll()
	for _, one := range ones {
		serverIds = append(serverIds, int64(one.(*Server).Id))
	}
	return
}

// 查找服务的搜索条件
func (this *ServerDAO) FindServerNodeFilters(serverId int64) (isOk bool, clusterId int64, err error) {
	one, err := this.Query().
		Pk(serverId).
		Result("clusterId").
		Find()
	if err != nil {
		return false, 0, err
	}
	if one == nil {
		isOk = false
		return
	}
	server := one.(*Server)
	return true, int64(server.ClusterId), nil
}

// 构造服务的Config
func (this *ServerDAO) ComposeServerConfig(serverId int64) (*serverconfigs.ServerConfig, error) {
	server, err := this.FindEnabledServer(serverId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("server not found")
	}

	config := &serverconfigs.ServerConfig{}
	config.Id = serverId
	config.Type = server.Type
	config.IsOn = server.IsOn == 1
	config.Name = server.Name
	config.Description = server.Description

	// ServerNames
	if len(server.ServerNames) > 0 && server.ServerNames != "null" {
		serverNames := []*serverconfigs.ServerNameConfig{}
		err = json.Unmarshal([]byte(server.ServerNames), &serverNames)
		if err != nil {
			return nil, err
		}
		config.ServerNames = serverNames
	}

	// HTTP
	if len(server.Http) > 0 && server.Http != "null" {
		httpConfig := &serverconfigs.HTTPProtocolConfig{}
		err = json.Unmarshal([]byte(server.Http), httpConfig)
		if err != nil {
			return nil, err
		}
		config.HTTP = httpConfig
	}

	// HTTPS
	if len(server.Https) > 0 && server.Https != "null" {
		httpsConfig := &serverconfigs.HTTPSProtocolConfig{}
		err = json.Unmarshal([]byte(server.Https), httpsConfig)
		if err != nil {
			return nil, err
		}

		// SSL
		if httpsConfig.SSLPolicyRef != nil && httpsConfig.SSLPolicyRef.SSLPolicyId > 0 {
			sslPolicyConfig, err := SharedSSLPolicyDAO.ComposePolicyConfig(httpsConfig.SSLPolicyRef.SSLPolicyId)
			if err != nil {
				return nil, err
			}
			if sslPolicyConfig != nil {
				httpsConfig.SSLPolicy = sslPolicyConfig
			}
		}

		config.HTTPS = httpsConfig
	}

	// TCP
	if len(server.Tcp) > 0 && server.Tcp != "null" {
		tcpConfig := &serverconfigs.TCPProtocolConfig{}
		err = json.Unmarshal([]byte(server.Tcp), tcpConfig)
		if err != nil {
			return nil, err
		}
		config.TCP = tcpConfig
	}

	// TLS
	if len(server.Tls) > 0 && server.Tls != "null" {
		tlsConfig := &serverconfigs.TLSProtocolConfig{}
		err = json.Unmarshal([]byte(server.Tls), tlsConfig)
		if err != nil {
			return nil, err
		}

		// SSL
		if tlsConfig.SSLPolicyRef != nil {
			sslPolicyConfig, err := SharedSSLPolicyDAO.ComposePolicyConfig(tlsConfig.SSLPolicyRef.SSLPolicyId)
			if err != nil {
				return nil, err
			}
			if sslPolicyConfig != nil {
				tlsConfig.SSLPolicy = sslPolicyConfig
			}
		}

		config.TLS = tlsConfig
	}

	// Unix
	if len(server.Unix) > 0 && server.Unix != "null" {
		unixConfig := &serverconfigs.UnixProtocolConfig{}
		err = json.Unmarshal([]byte(server.Unix), unixConfig)
		if err != nil {
			return nil, err
		}
		config.Unix = unixConfig
	}

	// UDP
	if len(server.Udp) > 0 && server.Udp != "null" {
		udpConfig := &serverconfigs.UDPProtocolConfig{}
		err = json.Unmarshal([]byte(server.Udp), udpConfig)
		if err != nil {
			return nil, err
		}
		config.UDP = udpConfig
	}

	// Web
	if server.WebId > 0 {
		webConfig, err := SharedHTTPWebDAO.ComposeWebConfig(int64(server.WebId))
		if err != nil {
			return nil, err
		}
		if webConfig != nil {
			config.Web = webConfig
		}
	}

	// ReverseProxy
	if IsNotNull(server.ReverseProxy) {
		reverseProxyRef := &serverconfigs.ReverseProxyRef{}
		err = json.Unmarshal([]byte(server.ReverseProxy), reverseProxyRef)
		if err != nil {
			return nil, err
		}
		config.ReverseProxyRef = reverseProxyRef

		reverseProxyConfig, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(reverseProxyRef.ReverseProxyId)
		if err != nil {
			return nil, err
		}
		if reverseProxyConfig != nil {
			config.ReverseProxy = reverseProxyConfig
		}
	}

	return config, nil
}

// 更新服务的Config配置
func (this *ServerDAO) RenewServerConfig(serverId int64, updateMd5 bool) (isChanged bool, err error) {
	serverConfig, err := this.ComposeServerConfig(serverId)
	if err != nil {
		return false, err
	}
	data, err := json.Marshal(serverConfig)
	if err != nil {
		return false, err
	}
	return this.UpdateServerConfig(serverId, data, updateMd5)
}

// 根据条件获取反向代理配置
func (this *ServerDAO) FindReverseProxyRef(serverId int64) (*serverconfigs.ReverseProxyRef, error) {
	reverseProxy, err := this.Query().
		Pk(serverId).
		Result("reverseProxy").
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

// 查找Server对应的WebId
func (this *ServerDAO) FindServerWebId(serverId int64) (int64, error) {
	webId, err := this.Query().
		Pk(serverId).
		Result("webId").
		FindIntCol(0)
	if err != nil {
		return 0, err
	}
	return int64(webId), nil
}

// 计算使用SSL策略的所有服务数量
func (this *ServerDAO) CountAllEnabledServersWithSSLPolicyIds(sslPolicyIds []int64) (count int64, err error) {
	if len(sslPolicyIds) == 0 {
		return
	}
	policyStringIds := []string{}
	for _, policyId := range sslPolicyIds {
		policyStringIds = append(policyStringIds, strconv.FormatInt(policyId, 10))
	}
	return this.Query().
		State(ServerStateEnabled).
		Where("(FIND_IN_SET(JSON_EXTRACT(https, '$.sslPolicyRef.sslPolicyId'), :policyIds) OR FIND_IN_SET(JSON_EXTRACT(tls, '$.sslPolicyRef.sslPolicyId'), :policyIds))").
		Param("policyIds", strings.Join(policyStringIds, ",")).
		Count()
}

// 查找使用某个SSL策略的所有服务
func (this *ServerDAO) FindAllEnabledServersWithSSLPolicyIds(sslPolicyIds []int64) (result []*Server, err error) {
	if len(sslPolicyIds) == 0 {
		return
	}
	policyStringIds := []string{}
	for _, policyId := range sslPolicyIds {
		policyStringIds = append(policyStringIds, strconv.FormatInt(policyId, 10))
	}
	_, err = this.Query().
		State(ServerStateEnabled).
		Result("id", "name", "https", "tls", "isOn", "type").
		Where("(FIND_IN_SET(JSON_EXTRACT(https, '$.sslPolicyRef.sslPolicyId'), :policyIds) OR FIND_IN_SET(JSON_EXTRACT(tls, '$.sslPolicyRef.sslPolicyId'), :policyIds))").
		Param("policyIds", strings.Join(policyStringIds, ",")).
		Slice(&result).
		AscPk().
		FindAll()
	return
}

// 计算使用某个缓存策略的所有服务数量
func (this *ServerDAO) CountEnabledServersWithWebIds(webIds []int64) (count int64, err error) {
	if len(webIds) == 0 {
		return
	}
	return this.Query().
		State(ServerStateEnabled).
		Attr("webId", webIds).
		Count()
}

// 查找使用某个缓存策略的所有服务
func (this *ServerDAO) FindAllEnabledServersWithWebIds(webIds []int64) (result []*Server, err error) {
	if len(webIds) == 0 {
		return
	}
	_, err = this.Query().
		State(ServerStateEnabled).
		Attr("webId", webIds).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 计算使用某个集群的所有服务数量
func (this *ServerDAO) CountAllEnabledServersWithNodeClusterId(clusterId int64) (int64, error) {
	return this.Query().
		State(ServerStateEnabled).
		Attr("clusterId", clusterId).
		Count()
}

// 计算使用某个分组的服务数量
func (this *ServerDAO) CountAllEnabledServersWithGroupId(groupId int64) (int64, error) {
	return this.Query().
		State(ServerStateEnabled).
		Where("JSON_CONTAINS(groupIds, :groupId)").
		Param("groupId", numberutils.FormatInt64(groupId)).
		Count()
}

// 创建事件
func (this *ServerDAO) createEvent() error {
	return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
}

// 生成DNS Name
func (this *ServerDAO) genDNSName() (string, error) {
	for {
		dnsName := rands.HexString(8)
		exist, err := this.Query().
			Attr("dnsName", dnsName).
			Exist()
		if err != nil {
			return "", err
		}
		if !exist {
			return dnsName, nil
		}
	}
}
