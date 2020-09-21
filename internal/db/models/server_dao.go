package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
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

var SharedServerDAO = NewServerDAO()

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
func (this *ServerDAO) CreateServer(adminId int64, userId int64, serverType serverconfigs.ServerType, name string, description string, serverNamesJSON string, httpJSON string, httpsJSON string, tcpJSON string, tlsJSON string, unixJSON string, udpJSON string, webId int64, reverseProxyJSON []byte, clusterId int64, includeNodesJSON string, excludeNodesJSON string) (serverId int64, err error) {
	uniqueId, err := this.genUniqueId()
	if err != nil {
		return 0, err
	}

	op := NewServerOperator()
	op.UniqueId = uniqueId
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
	op.GroupIds = "[]"
	op.Version = 1
	op.IsOn = 1
	op.State = ServerStateEnabled
	_, err = this.Save(op)

	if err != nil {
		return 0, err
	}

	serverId = types.Int64(op.Id)
	err = this.RenewServerConfig(serverId)
	return serverId, err
}

// 修改服务基本信息
func (this *ServerDAO) UpdateServerBasic(serverId int64, name string, description string, clusterId int64) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	op := NewServerOperator()
	op.Id = serverId
	op.Name = name
	op.Description = description
	op.ClusterId = clusterId
	op.Version = dbs.SQL("version=version+1")
	_, err := this.Save(op)
	return err
}

// 修改服务配置
func (this *ServerDAO) UpdateServerConfig(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("config", string(config)).
		Set("version", dbs.SQL("version+1")).
		Update()
	return err
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
	return this.RenewServerConfig(serverId)
}

// 修改HTTPS配置
func (this *ServerDAO) UpdateServerHTTPS(serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query().
		Pk(serverId).
		Set("https", string(config)).
		Update()
	if err != nil {
		return err
	}
	return this.RenewServerConfig(serverId)
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
	return this.RenewServerConfig(serverId)
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
	return this.RenewServerConfig(serverId)
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
	return this.RenewServerConfig(serverId)
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
	return this.RenewServerConfig(serverId)
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
	return this.RenewServerConfig(serverId)
}

// 初始化Web配置
func (this *ServerDAO) InitServerWeb(serverId int64) (int64, error) {
	if serverId <= 0 {
		return 0, errors.New("serverId should not be smaller than 0")
	}

	webId, err := SharedHTTPWebDAO.CreateWeb("")
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
	return this.RenewServerConfig(serverId)
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
	return this.RenewServerConfig(serverId)
}

// 计算所有可用服务数量
func (this *ServerDAO) CountAllEnabledServers() (int64, error) {
	return this.Query().
		State(ServerStateEnabled).
		Count()
}

// 列出单页的服务
func (this *ServerDAO) ListEnabledServers(offset int64, size int64) (result []*Server, err error) {
	_, err = this.Query().
		State(ServerStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
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

	// Components
	// TODO

	// Filters
	// TODO

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
func (this *ServerDAO) RenewServerConfig(serverId int64) error {
	serverConfig, err := this.ComposeServerConfig(serverId)
	if err != nil {
		return err
	}
	data, err := serverConfig.AsJSON()
	if err != nil {
		return err
	}
	return this.UpdateServerConfig(serverId, data)
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

// 查找需要更新的Server
func (this *ServerDAO) FindUpdatingServerIds() (serverIds []int64, err error) {
	ones, err := this.Query().
		State(ServerStateEnabled).
		Attr("isUpdating", true).
		ResultPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		serverIds = append(serverIds, int64(one.(*Server).Id))
	}
	return
}

// 修改服务是否需要更新
func (this *ServerDAO) UpdateServerIsUpdating(serverId int64, isUpdating bool) error {
	_, err := this.Query().
		Pk(serverId).
		Set("isUpdating", isUpdating).
		Update()
	return err
}

// 查找WebId
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

// 更新所有Web相关的处于更新状态
func (this *ServerDAO) UpdateServerIsUpdatingWithWebId(webId int64) error {
	_, err := this.Query().
		Attr("webId", webId).
		Set("isUpdating", true).
		Update()
	return err
}

// 生成唯一ID
func (this *ServerDAO) genUniqueId() (string, error) {
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
