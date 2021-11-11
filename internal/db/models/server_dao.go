package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"strconv"
	"strings"
	"time"
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

// Init 初始化
func (this *ServerDAO) Init() {
	_ = this.DAOObject.Init()

	// 这里不处理增删改事件，是为了避免Server修改本身的时候，也要触发别的Server变更
}

// EnableServer 启用条目
func (this *ServerDAO) EnableServer(tx *dbs.Tx, id uint32) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", ServerStateEnabled).
		Update()
}

// DisableServer 禁用条目
func (this *ServerDAO) DisableServer(tx *dbs.Tx, serverId int64) (err error) {
	_, err = this.Query(tx).
		Pk(serverId).
		Set("state", ServerStateDisabled).
		Update()
	if err != nil {
		return err
	}
	err = this.NotifyUpdate(tx, serverId)
	if err != nil {
		return err
	}

	err = this.NotifyDNSUpdate(tx, serverId)
	if err != nil {
		return err
	}
	return nil
}

// FindEnabledServer 查找启用中的服务
func (this *ServerDAO) FindEnabledServer(tx *dbs.Tx, serverId int64) (*Server, error) {
	result, err := this.Query(tx).
		Pk(serverId).
		Attr("state", ServerStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Server), err
}

// FindEnabledServerName 查找服务名称
func (this *ServerDAO) FindEnabledServerName(tx *dbs.Tx, serverId int64) (string, error) {
	return this.Query(tx).
		Pk(serverId).
		State(ServerStateEnabled).
		Result("name").
		FindStringCol("")
}

// FindEnabledServerBasic 查找服务基本信息
func (this *ServerDAO) FindEnabledServerBasic(tx *dbs.Tx, serverId int64) (*Server, error) {
	result, err := this.Query(tx).
		Pk(serverId).
		State(ServerStateEnabled).
		Result("id", "name", "description", "isOn", "type", "clusterId").
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Server), err
}

// FindEnabledServerType 查找服务类型
func (this *ServerDAO) FindEnabledServerType(tx *dbs.Tx, serverId int64) (string, error) {
	return this.Query(tx).
		Pk(serverId).
		Result("type").
		FindStringCol("")
}

// CreateServer 创建服务
func (this *ServerDAO) CreateServer(tx *dbs.Tx,
	adminId int64,
	userId int64,
	serverType serverconfigs.ServerType,
	name string,
	description string,
	serverNamesJSON []byte,
	isAuditing bool,
	auditingServerNamesJSON []byte,
	httpJSON string,
	httpsJSON string,
	tcpJSON string,
	tlsJSON string,
	unixJSON string,
	udpJSON string,
	webId int64,
	reverseProxyJSON []byte,
	clusterId int64,
	includeNodesJSON string,
	excludeNodesJSON string,
	groupIds []int64,
	userPlanId int64) (serverId int64, err error) {
	op := NewServerOperator()
	op.UserId = userId
	op.AdminId = adminId
	op.Name = name
	op.Type = serverType
	op.Description = description

	if len(serverNamesJSON) > 0 {
		op.ServerNames = serverNamesJSON
	}
	op.IsAuditing = isAuditing
	if len(auditingServerNamesJSON) > 0 {
		op.AuditingServerNames = auditingServerNamesJSON
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

	dnsName, err := this.GenDNSName(tx)
	if err != nil {
		return 0, err
	}
	op.DnsName = dnsName

	op.UserPlanId = userPlanId

	op.Version = 1
	op.IsOn = 1
	op.State = ServerStateEnabled
	err = this.Save(tx, op)

	if err != nil {
		return 0, err
	}

	serverId = types.Int64(op.Id)

	// 更新端口
	err = this.NotifyServerPortsUpdate(tx, serverId)
	if err != nil {
		return serverId, err
	}

	// 通知配置更改
	err = this.NotifyUpdate(tx, serverId)
	if err != nil {
		return 0, err
	}

	// 通知DNS更改
	err = this.NotifyDNSUpdate(tx, serverId)
	if err != nil {
		return 0, err
	}

	return serverId, nil
}

// UpdateServerBasic 修改服务基本信息
func (this *ServerDAO) UpdateServerBasic(tx *dbs.Tx, serverId int64, name string, description string, clusterId int64, isOn bool, groupIds []int64) error {
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

	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	// 通知更新
	err = this.NotifyUpdate(tx, serverId)
	if err != nil {
		return err
	}

	// 因为可能有isOn的原因，所以需要修改
	return this.NotifyDNSUpdate(tx, serverId)
}

// UpdateUserServerBasic 设置用户相关的基本信息
func (this *ServerDAO) UpdateUserServerBasic(tx *dbs.Tx, serverId int64, name string) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	op := NewServerOperator()
	op.Id = serverId
	op.Name = name

	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerIsOn 修复服务是否启用
func (this *ServerDAO) UpdateServerIsOn(tx *dbs.Tx, serverId int64, isOn bool) error {
	_, err := this.Query(tx).
		Pk(serverId).
		Set("isOn", isOn).
		Update()
	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, serverId)
	if err != nil {
		return err
	}

	return nil
}

// UpdateServerHTTP 修改HTTP配置
func (this *ServerDAO) UpdateServerHTTP(tx *dbs.Tx, serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query(tx).
		Pk(serverId).
		Set("http", string(config)).
		Update()
	if err != nil {
		return err
	}

	// 更新端口
	err = this.NotifyServerPortsUpdate(tx, serverId)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerHTTPS 修改HTTPS配置
func (this *ServerDAO) UpdateServerHTTPS(tx *dbs.Tx, serverId int64, httpsJSON []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(httpsJSON) == 0 {
		httpsJSON = []byte("null")
	}
	_, err := this.Query(tx).
		Pk(serverId).
		Set("https", string(httpsJSON)).
		Update()
	if err != nil {
		return err
	}

	// 更新端口
	err = this.NotifyServerPortsUpdate(tx, serverId)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerTCP 修改TCP配置
func (this *ServerDAO) UpdateServerTCP(tx *dbs.Tx, serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query(tx).
		Pk(serverId).
		Set("tcp", string(config)).
		Update()
	if err != nil {
		return err
	}

	// 更新端口
	err = this.NotifyServerPortsUpdate(tx, serverId)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerTLS 修改TLS配置
func (this *ServerDAO) UpdateServerTLS(tx *dbs.Tx, serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query(tx).
		Pk(serverId).
		Set("tls", string(config)).
		Update()
	if err != nil {
		return err
	}

	// 更新端口
	err = this.NotifyServerPortsUpdate(tx, serverId)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerUnix 修改Unix配置
func (this *ServerDAO) UpdateServerUnix(tx *dbs.Tx, serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query(tx).
		Pk(serverId).
		Set("unix", string(config)).
		Update()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerUDP 修改UDP配置
func (this *ServerDAO) UpdateServerUDP(tx *dbs.Tx, serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	if len(config) == 0 {
		config = []byte("null")
	}
	_, err := this.Query(tx).
		Pk(serverId).
		Set("udp", string(config)).
		Update()
	if err != nil {
		return err
	}

	// 更新端口
	err = this.NotifyServerPortsUpdate(tx, serverId)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerWeb 修改Web配置
func (this *ServerDAO) UpdateServerWeb(tx *dbs.Tx, serverId int64, webId int64) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	_, err := this.Query(tx).
		Pk(serverId).
		Set("webId", webId).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerDNS 修改DNS设置
func (this *ServerDAO) UpdateServerDNS(tx *dbs.Tx, serverId int64, supportCNAME bool) error {
	if serverId <= 0 {
		return errors.New("invalid serverId")
	}
	var op = NewServerOperator()
	op.Id = serverId
	op.SupportCNAME = supportCNAME
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// InitServerWeb 初始化Web配置
func (this *ServerDAO) InitServerWeb(tx *dbs.Tx, serverId int64) (int64, error) {
	if serverId <= 0 {
		return 0, errors.New("invalid serverId")
	}

	adminId, userId, err := this.FindServerAdminIdAndUserId(tx, serverId)
	if err != nil {
		return 0, err
	}

	webId, err := SharedHTTPWebDAO.CreateWeb(tx, adminId, userId, nil)
	if err != nil {
		return 0, err
	}

	_, err = this.Query(tx).
		Pk(serverId).
		Set("webId", webId).
		Update()
	if err != nil {
		return 0, err
	}

	err = this.NotifyUpdate(tx, serverId)
	if err != nil {
		return webId, err
	}

	return webId, nil
}

// FindServerServerNames 查找ServerNames配置
func (this *ServerDAO) FindServerServerNames(tx *dbs.Tx, serverId int64) (serverNamesJSON []byte, isAuditing bool, auditingServerNamesJSON []byte, auditingResultJSON []byte, err error) {
	if serverId <= 0 {
		return
	}
	one, err := this.Query(tx).
		Pk(serverId).
		Result("serverNames", "isAuditing", "auditingServerNames", "auditingResult").
		Find()
	if err != nil {
		return nil, false, nil, nil, err
	}
	if one == nil {
		return
	}
	server := one.(*Server)
	return []byte(server.ServerNames), server.IsAuditing == 1, []byte(server.AuditingServerNames), []byte(server.AuditingResult), nil
}

// UpdateServerNames 修改ServerNames配置
func (this *ServerDAO) UpdateServerNames(tx *dbs.Tx, serverId int64, serverNames []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}

	op := NewServerOperator()
	op.Id = serverId

	if len(serverNames) == 0 {
		serverNames = []byte("[]")
	}
	op.ServerNames = serverNames
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, serverId)
}

// UpdateAuditingServerNames 修改域名审核
func (this *ServerDAO) UpdateAuditingServerNames(tx *dbs.Tx, serverId int64, isAuditing bool, auditingServerNamesJSON []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}

	op := NewServerOperator()
	op.Id = serverId
	op.IsAuditing = isAuditing
	if len(auditingServerNamesJSON) == 0 {
		op.AuditingServerNames = "[]"
	} else {
		op.AuditingServerNames = auditingServerNamesJSON
	}
	op.AuditingResult = `{"isOk":true}`
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, serverId)
}

// UpdateServerAuditing 修改域名审核结果
func (this *ServerDAO) UpdateServerAuditing(tx *dbs.Tx, serverId int64, result *pb.ServerNameAuditingResult) error {
	if serverId <= 0 {
		return errors.New("invalid serverId")
	}

	resultJSON, err := json.Marshal(maps.Map{
		"isOk":      result.IsOk,
		"reason":    result.Reason,
		"createdAt": time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	op := NewServerOperator()
	op.Id = serverId
	op.IsAuditing = false
	op.AuditingResult = resultJSON
	if result.IsOk {
		op.ServerNames = dbs.SQL("auditingServerNames")
	}
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	err = this.NotifyUpdate(tx, serverId)
	if err != nil {
		return err
	}

	return this.NotifyDNSUpdate(tx, serverId)
}

// UpdateServerReverseProxy 修改反向代理配置
func (this *ServerDAO) UpdateServerReverseProxy(tx *dbs.Tx, serverId int64, config []byte) error {
	if serverId <= 0 {
		return errors.New("serverId should not be smaller than 0")
	}
	op := NewServerOperator()
	op.Id = serverId
	op.ReverseProxy = JSONBytes(config)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, serverId)
}

// CountAllEnabledServers 计算所有可用服务数量
func (this *ServerDAO) CountAllEnabledServers(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(ServerStateEnabled).
		Count()
}

// CountAllEnabledServersMatch 计算所有可用服务数量
func (this *ServerDAO) CountAllEnabledServersMatch(tx *dbs.Tx, groupId int64, keyword string, userId int64, clusterId int64, auditingFlag configutils.BoolState, protocolFamily string) (int64, error) {
	query := this.Query(tx).
		State(ServerStateEnabled)
	if groupId > 0 {
		query.Where("JSON_CONTAINS(groupIds, :groupId)").
			Param("groupId", numberutils.FormatInt64(groupId))
	}
	if len(keyword) > 0 {
		if regexp.MustCompile(`^\d+$`).MatchString(keyword) {
			query.Where("(name LIKE :keyword OR serverNames LIKE :keyword OR JSON_CONTAINS(http, :portRange, '$.listen') OR JSON_CONTAINS(https, :portRange, '$.listen') OR JSON_CONTAINS(tcp, :portRange, '$.listen') OR JSON_CONTAINS(tls, :portRange, '$.listen'))").
				Param("portRange", maps.Map{"portRange": keyword}.AsJSON()).
				Param("keyword", "%"+keyword+"%")
		} else {
			query.Where("(name LIKE :keyword OR serverNames LIKE :keyword)").
				Param("keyword", "%"+keyword+"%")
		}
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if auditingFlag == configutils.BoolStateYes {
		query.Attr("isAuditing", true)
	}
	if protocolFamily == "http" {
		query.Where("(http IS NOT NULL OR https IS NOT NULL)")
	} else if protocolFamily == "tcp" {
		query.Where("(tcp IS NOT NULL OR tls IS NOT NULL)")
	}
	return query.Count()
}

// ListEnabledServersMatch 列出单页的服务
func (this *ServerDAO) ListEnabledServersMatch(tx *dbs.Tx, offset int64, size int64, groupId int64, keyword string, userId int64, clusterId int64, auditingFlag int32, protocolFamily string) (result []*Server, err error) {
	query := this.Query(tx).
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
		if regexp.MustCompile(`^\d+$`).MatchString(keyword) {
			query.Where("(name LIKE :keyword OR serverNames LIKE :keyword OR JSON_CONTAINS(http, :portRange, '$.listen') OR JSON_CONTAINS(https, :portRange, '$.listen') OR JSON_CONTAINS(tcp, :portRange, '$.listen') OR JSON_CONTAINS(tls, :portRange, '$.listen'))").
				Param("portRange", string(maps.Map{"portRange": keyword}.AsJSON())).
				Param("keyword", "%"+keyword+"%")
		} else {
			query.Where("(name LIKE :keyword OR serverNames LIKE :keyword)").
				Param("keyword", "%"+keyword+"%")
		}
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if auditingFlag == 1 {
		query.Attr("isAuditing", true)
	}
	if protocolFamily == "http" {
		query.Where("(http IS NOT NULL OR https IS NOT NULL)")
	} else if protocolFamily == "tcp" {
		query.Where("(tcp IS NOT NULL OR tls IS NOT NULL)")
	}

	_, err = query.FindAll()
	return
}

// FindAllEnabledServersWithNode 获取节点中的所有服务
func (this *ServerDAO) FindAllEnabledServersWithNode(tx *dbs.Tx, nodeId int64) (result []*Server, err error) {
	// 节点所在主集群
	clusterIds, err := SharedNodeDAO.FindEnabledAndOnNodeClusterIds(tx, nodeId)
	if err != nil {
		return nil, err
	}
	for _, clusterId := range clusterIds {
		ones, err := this.Query(tx).
			Attr("clusterId", clusterId).
			State(ServerStateEnabled).
			AscPk().
			FindAll()
		if err != nil {
			return nil, err
		}
		for _, one := range ones {
			result = append(result, one.(*Server))
		}
	}
	return
}

// FindAllEnabledServerIds 获取所有的服务ID
func (this *ServerDAO) FindAllEnabledServerIds(tx *dbs.Tx) (serverIds []int64, err error) {
	ones, err := this.Query(tx).
		State(ServerStateEnabled).
		AscPk().
		ResultPk().
		FindAll()
	for _, one := range ones {
		serverIds = append(serverIds, int64(one.(*Server).Id))
	}
	return
}

// FindAllEnabledServerIdsWithUserId 获取某个用户的所有的服务ID
func (this *ServerDAO) FindAllEnabledServerIdsWithUserId(tx *dbs.Tx, userId int64) (serverIds []int64, err error) {
	ones, err := this.Query(tx).
		State(ServerStateEnabled).
		Attr("userId", userId).
		AscPk().
		ResultPk().
		FindAll()
	for _, one := range ones {
		serverIds = append(serverIds, int64(one.(*Server).Id))
	}
	return
}

// FindAllEnabledServerIdsWithGroupId 获取某个分组下的所有的服务ID
func (this *ServerDAO) FindAllEnabledServerIdsWithGroupId(tx *dbs.Tx, groupId int64) (serverIds []int64, err error) {
	ones, err := this.Query(tx).
		State(ServerStateEnabled).
		Where("JSON_CONTAINS(groupIds, :groupId)").
		Param("groupId", numberutils.FormatInt64(groupId)).
		AscPk().
		ResultPk().
		FindAll()
	for _, one := range ones {
		serverIds = append(serverIds, int64(one.(*Server).Id))
	}
	return
}

// FindServerGroupIds 获取服务的分组ID
func (this *ServerDAO) FindServerGroupIds(tx *dbs.Tx, serverId int64) ([]int64, error) {
	if serverId <= 0 {
		return nil, nil
	}
	groupIdsString, err := this.Query(tx).
		Pk(serverId).
		Result("groupIds").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(groupIdsString) == 0 {
		return nil, nil
	}
	var result = []int64{}
	err = json.Unmarshal([]byte(groupIdsString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// FindServerNodeFilters 查找服务的搜索条件
func (this *ServerDAO) FindServerNodeFilters(tx *dbs.Tx, serverId int64) (isOk bool, clusterId int64, err error) {
	one, err := this.Query(tx).
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

// ComposeServerConfigWithServerId 构造服务的Config
func (this *ServerDAO) ComposeServerConfigWithServerId(tx *dbs.Tx, serverId int64, forNode bool) (*serverconfigs.ServerConfig, error) {
	server, err := this.FindEnabledServer(tx, serverId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, ErrNotFound
	}
	return this.ComposeServerConfig(tx, server, nil, forNode)
}

// ComposeServerConfig 构造服务的Config
// forNode 是否是节点请求
func (this *ServerDAO) ComposeServerConfig(tx *dbs.Tx, server *Server, cacheMap maps.Map, forNode bool) (*serverconfigs.ServerConfig, error) {
	if server == nil {
		return nil, ErrNotFound
	}

	if cacheMap == nil {
		cacheMap = maps.Map{}
	}

	config := &serverconfigs.ServerConfig{}
	config.Id = int64(server.Id)
	config.ClusterId = int64(server.ClusterId)
	config.Type = server.Type
	config.IsOn = server.IsOn == 1
	config.Name = server.Name
	config.Description = server.Description

	var groupConfig *serverconfigs.ServerGroupConfig
	for _, groupId := range server.DecodeGroupIds() {
		groupConfig1, err := SharedServerGroupDAO.ComposeGroupConfig(tx, groupId, cacheMap)
		if err != nil {
			return nil, err
		}
		if groupConfig1 == nil {
			continue
		}
		groupConfig = groupConfig1
		break
	}
	config.Group = groupConfig

	// ServerNames
	if len(server.ServerNames) > 0 && server.ServerNames != "null" {
		serverNames := []*serverconfigs.ServerNameConfig{}
		err := json.Unmarshal([]byte(server.ServerNames), &serverNames)
		if err != nil {
			return nil, err
		}
		config.ServerNames = serverNames
	}

	// CNAME
	config.SupportCNAME = server.SupportCNAME == 1
	if server.ClusterId > 0 && len(server.DnsName) > 0 {
		clusterDNS, err := SharedNodeClusterDAO.FindClusterDNSInfo(tx, int64(server.ClusterId), cacheMap)
		if err != nil {
			return nil, err
		}
		if clusterDNS != nil && clusterDNS.DnsDomainId > 0 {
			domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, int64(clusterDNS.DnsDomainId), cacheMap)
			if err != nil {
				return nil, err
			}
			if domain != nil {
				cname := server.DnsName + "." + domain.Name
				config.AliasServerNames = append(config.AliasServerNames, cname)
			}
		}
	}

	// HTTP
	if len(server.Http) > 0 && server.Http != "null" {
		httpConfig := &serverconfigs.HTTPProtocolConfig{}
		err := json.Unmarshal([]byte(server.Http), httpConfig)
		if err != nil {
			return nil, err
		}
		config.HTTP = httpConfig
	}

	// HTTPS
	if len(server.Https) > 0 && server.Https != "null" {
		httpsConfig := &serverconfigs.HTTPSProtocolConfig{}
		err := json.Unmarshal([]byte(server.Https), httpsConfig)
		if err != nil {
			return nil, err
		}

		// SSL
		if httpsConfig.SSLPolicyRef != nil && httpsConfig.SSLPolicyRef.SSLPolicyId > 0 {
			sslPolicyConfig, err := SharedSSLPolicyDAO.ComposePolicyConfig(tx, httpsConfig.SSLPolicyRef.SSLPolicyId, cacheMap)
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
		err := json.Unmarshal([]byte(server.Tcp), tcpConfig)
		if err != nil {
			return nil, err
		}
		config.TCP = tcpConfig
	}

	// TLS
	if len(server.Tls) > 0 && server.Tls != "null" {
		tlsConfig := &serverconfigs.TLSProtocolConfig{}
		err := json.Unmarshal([]byte(server.Tls), tlsConfig)
		if err != nil {
			return nil, err
		}

		// SSL
		if tlsConfig.SSLPolicyRef != nil {
			sslPolicyConfig, err := SharedSSLPolicyDAO.ComposePolicyConfig(tx, tlsConfig.SSLPolicyRef.SSLPolicyId, cacheMap)
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
		err := json.Unmarshal([]byte(server.Unix), unixConfig)
		if err != nil {
			return nil, err
		}
		config.Unix = unixConfig
	}

	// UDP
	if len(server.Udp) > 0 && server.Udp != "null" {
		udpConfig := &serverconfigs.UDPProtocolConfig{}
		err := json.Unmarshal([]byte(server.Udp), udpConfig)
		if err != nil {
			return nil, err
		}
		config.UDP = udpConfig
	}

	// Web
	if server.WebId > 0 {
		webConfig, err := SharedHTTPWebDAO.ComposeWebConfig(tx, int64(server.WebId), cacheMap)
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
		err := json.Unmarshal([]byte(server.ReverseProxy), reverseProxyRef)
		if err != nil {
			return nil, err
		}
		config.ReverseProxyRef = reverseProxyRef

		reverseProxyConfig, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, cacheMap)
		if err != nil {
			return nil, err
		}
		if reverseProxyConfig != nil {
			config.ReverseProxy = reverseProxyConfig
		}
	}

	// WAF策略
	clusterId := int64(server.ClusterId)
	httpFirewallPolicyId, err := SharedNodeClusterDAO.FindClusterHTTPFirewallPolicyId(tx, clusterId, cacheMap)
	if err != nil {
		return nil, err
	}
	if httpFirewallPolicyId > 0 {
		config.HTTPFirewallPolicyId = httpFirewallPolicyId
	}

	// 缓存策略
	httpCachePolicyId, err := SharedNodeClusterDAO.FindClusterHTTPCachePolicyId(tx, clusterId, cacheMap)
	if err != nil {
		return nil, err
	}
	if httpCachePolicyId > 0 {
		config.HTTPCachePolicyId = httpCachePolicyId
	}

	// traffic limit
	if len(server.TrafficLimit) > 0 {
		var trafficLimitConfig = &serverconfigs.TrafficLimitConfig{}
		err = json.Unmarshal([]byte(server.TrafficLimit), trafficLimitConfig)
		if err != nil {
			return nil, err
		}
		config.TrafficLimit = trafficLimitConfig
	}

	// 用户套餐
	if forNode && server.UserPlanId > 0 {
		userPlan, err := SharedUserPlanDAO.FindEnabledUserPlan(tx, int64(server.UserPlanId), cacheMap)
		if err != nil {
			return nil, err
		}
		if userPlan != nil && userPlan.IsOn == 1 {
			if len(userPlan.DayTo) == 0 {
				userPlan.DayTo = DefaultUserPlanMaxDay
			}

			// 套餐是否依然有效
			plan, err := SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
			if err != nil {
				return nil, err
			}
			if plan != nil {
				config.UserPlan = &serverconfigs.UserPlanConfig{
					DayTo: userPlan.DayTo,
					Plan: &serverconfigs.PlanConfig{
						Id: int64(plan.Id),
					},
				}

				if len(plan.TrafficLimit) > 0 && (config.TrafficLimit == nil || !config.TrafficLimit.IsOn) {
					var trafficLimitConfig = &serverconfigs.TrafficLimitConfig{}
					err = json.Unmarshal([]byte(plan.TrafficLimit), trafficLimitConfig)
					if err != nil {
						return nil, err
					}
					config.TrafficLimit = trafficLimitConfig
				}
			}
		}
	}

	if config.TrafficLimit != nil && config.TrafficLimit.IsOn && !config.TrafficLimit.IsEmpty() {
		if len(server.TrafficLimitStatus) > 0 {
			var status = &serverconfigs.TrafficLimitStatus{}
			err = json.Unmarshal([]byte(server.TrafficLimitStatus), status)
			if err != nil {
				return nil, err
			}
			if status.IsValid() {
				config.TrafficLimitStatus = status
			}
		}
	}

	return config, nil
}

// FindReverseProxyRef 根据条件获取反向代理配置
func (this *ServerDAO) FindReverseProxyRef(tx *dbs.Tx, serverId int64) (*serverconfigs.ReverseProxyRef, error) {
	reverseProxy, err := this.Query(tx).
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

// FindServerWebId 查找Server对应的WebId
func (this *ServerDAO) FindServerWebId(tx *dbs.Tx, serverId int64) (int64, error) {
	webId, err := this.Query(tx).
		Pk(serverId).
		Result("webId").
		FindIntCol(0)
	if err != nil {
		return 0, err
	}
	return int64(webId), nil
}

// CountAllEnabledServersWithSSLPolicyIds 计算使用SSL策略的所有服务数量
func (this *ServerDAO) CountAllEnabledServersWithSSLPolicyIds(tx *dbs.Tx, sslPolicyIds []int64) (count int64, err error) {
	if len(sslPolicyIds) == 0 {
		return
	}
	policyStringIds := []string{}
	for _, policyId := range sslPolicyIds {
		policyStringIds = append(policyStringIds, strconv.FormatInt(policyId, 10))
	}
	return this.Query(tx).
		State(ServerStateEnabled).
		Where("(FIND_IN_SET(JSON_EXTRACT(https, '$.sslPolicyRef.sslPolicyId'), :policyIds) OR FIND_IN_SET(JSON_EXTRACT(tls, '$.sslPolicyRef.sslPolicyId'), :policyIds))").
		Param("policyIds", strings.Join(policyStringIds, ",")).
		Count()
}

// FindAllEnabledServersWithSSLPolicyIds 查找使用某个SSL策略的所有服务
func (this *ServerDAO) FindAllEnabledServersWithSSLPolicyIds(tx *dbs.Tx, sslPolicyIds []int64) (result []*Server, err error) {
	if len(sslPolicyIds) == 0 {
		return
	}
	policyStringIds := []string{}
	for _, policyId := range sslPolicyIds {
		policyStringIds = append(policyStringIds, strconv.FormatInt(policyId, 10))
	}
	_, err = this.Query(tx).
		State(ServerStateEnabled).
		Result("id", "name", "https", "tls", "isOn", "type").
		Where("(FIND_IN_SET(JSON_EXTRACT(https, '$.sslPolicyRef.sslPolicyId'), :policyIds) OR FIND_IN_SET(JSON_EXTRACT(tls, '$.sslPolicyRef.sslPolicyId'), :policyIds))").
		Param("policyIds", strings.Join(policyStringIds, ",")).
		Slice(&result).
		AscPk().
		FindAll()
	return
}

// FindAllEnabledServerIdsWithSSLPolicyIds 查找使用某个SSL策略的所有服务Id
func (this *ServerDAO) FindAllEnabledServerIdsWithSSLPolicyIds(tx *dbs.Tx, sslPolicyIds []int64) (result []int64, err error) {
	if len(sslPolicyIds) == 0 {
		return
	}

	for _, policyId := range sslPolicyIds {
		ones, err := this.Query(tx).
			State(ServerStateEnabled).
			ResultPk().
			Where("(JSON_CONTAINS(https, :jsonQuery) OR JSON_CONTAINS(tls, :jsonQuery))").
			Param("jsonQuery", maps.Map{"sslPolicyRef": maps.Map{"sslPolicyId": policyId}}.AsJSON()).
			FindAll()
		if err != nil {
			return nil, err
		}
		for _, one := range ones {
			serverId := int64(one.(*Server).Id)
			if !lists.ContainsInt64(result, serverId) {
				result = append(result, serverId)
			}
		}
	}
	return
}

// CountEnabledServersWithWebIds 计算使用某个缓存策略的所有服务数量
func (this *ServerDAO) CountEnabledServersWithWebIds(tx *dbs.Tx, webIds []int64) (count int64, err error) {
	if len(webIds) == 0 {
		return
	}
	return this.Query(tx).
		State(ServerStateEnabled).
		Attr("webId", webIds).
		Reuse(false).
		Count()
}

// FindAllEnabledServersWithWebIds 查找使用某个缓存策略的所有服务
func (this *ServerDAO) FindAllEnabledServersWithWebIds(tx *dbs.Tx, webIds []int64) (result []*Server, err error) {
	if len(webIds) == 0 {
		return
	}
	_, err = this.Query(tx).
		State(ServerStateEnabled).
		Attr("webId", webIds).
		Reuse(false).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledServersWithNodeClusterId 计算使用某个集群的所有服务数量
func (this *ServerDAO) CountAllEnabledServersWithNodeClusterId(tx *dbs.Tx, clusterId int64) (int64, error) {
	return this.Query(tx).
		State(ServerStateEnabled).
		Attr("clusterId", clusterId).
		Count()
}

// CountAllEnabledServersWithGroupId 计算使用某个分组的服务数量
func (this *ServerDAO) CountAllEnabledServersWithGroupId(tx *dbs.Tx, groupId int64) (int64, error) {
	return this.Query(tx).
		State(ServerStateEnabled).
		Where("JSON_CONTAINS(groupIds, :groupId)").
		Param("groupId", numberutils.FormatInt64(groupId)).
		Count()
}

// FindAllServerDNSNamesWithDNSDomainId 查询使用某个DNS域名的所有服务域名
func (this *ServerDAO) FindAllServerDNSNamesWithDNSDomainId(tx *dbs.Tx, dnsDomainId int64) ([]string, error) {
	clusterIds, err := SharedNodeClusterDAO.FindAllEnabledClusterIdsWithDNSDomainId(tx, dnsDomainId)
	if err != nil {
		return nil, err
	}
	if len(clusterIds) == 0 {
		return nil, nil
	}
	ones, err := this.Query(tx).
		State(ServerStateEnabled).
		Attr("isOn", true).
		Attr("clusterId", clusterIds).
		Result("dnsName").
		Reuse(false). // 避免因为IN语句造成内存占用过多
		FindAll()
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, one := range ones {
		dnsName := one.(*Server).DnsName
		if len(dnsName) == 0 {
			continue
		}
		result = append(result, dnsName)
	}
	return result, nil
}

// FindAllServersDNSWithClusterId 获取某个集群下的服务DNS信息
func (this *ServerDAO) FindAllServersDNSWithClusterId(tx *dbs.Tx, clusterId int64) (result []*Server, err error) {
	_, err = this.Query(tx).
		State(ServerStateEnabled).
		Attr("isOn", true).
		Attr("isAuditing", false). // 不在审核中
		Attr("clusterId", clusterId).
		Result("id", "name", "dnsName").
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledServersWithDomain 根据域名查找服务
func (this *ServerDAO) FindAllEnabledServersWithDomain(tx *dbs.Tx, domain string) (result []*Server, err error) {
	if len(domain) == 0 {
		return
	}

	_, err = this.Query(tx).
		State(ServerStateEnabled).
		Where("(JSON_CONTAINS(serverNames, :domain1) OR JSON_CONTAINS(serverNames, :domain2))").
		Param("domain1", maps.Map{"name": domain}.AsJSON()).
		Param("domain2", maps.Map{"subNames": domain}.AsJSON()).
		Slice(&result).
		DescPk().
		FindAll()

	if err != nil {
		return nil, err
	}

	// 支持泛解析
	var countPieces = strings.Count(domain, ".")
	for {
		var index = strings.Index(domain, ".")
		if index > 0 {
			domain = domain[index+1:]
			var search = strings.Repeat("*.", countPieces-strings.Count(domain, ".")) + domain
			_, err = this.Query(tx).
				State(ServerStateEnabled).
				Where("(JSON_CONTAINS(serverNames, :domain1) OR JSON_CONTAINS(serverNames, :domain2))").
				Param("domain1", maps.Map{"name": search}.AsJSON()).
				Param("domain2", maps.Map{"subNames": search}.AsJSON()).
				Slice(&result).
				DescPk().
				FindAll()
			if err != nil {
				return
			}
			if len(result) > 0 {
				return
			}
		} else {
			break
		}
	}

	return
}

// GenerateServerDNSName 重新生成子域名
func (this *ServerDAO) GenerateServerDNSName(tx *dbs.Tx, serverId int64) (string, error) {
	if serverId <= 0 {
		return "", errors.New("invalid serverId")
	}
	dnsName, err := this.GenDNSName(tx)
	if err != nil {
		return "", err
	}
	op := NewServerOperator()
	op.Id = serverId
	op.DnsName = dnsName
	err = this.Save(tx, op)
	if err != nil {
		return "", err
	}

	err = this.NotifyUpdate(tx, serverId)
	if err != nil {
		return "", err
	}

	err = this.NotifyDNSUpdate(tx, serverId)
	if err != nil {
		return "", err
	}

	return dnsName, nil
}

// FindServerClusterId 查询当前服务的集群ID
func (this *ServerDAO) FindServerClusterId(tx *dbs.Tx, serverId int64) (int64, error) {
	return this.Query(tx).
		Pk(serverId).
		Result("clusterId").
		FindInt64Col(0)
}

// FindServerDNSName 查询服务的DNS名称
func (this *ServerDAO) FindServerDNSName(tx *dbs.Tx, serverId int64) (string, error) {
	return this.Query(tx).
		Pk(serverId).
		Result("dnsName").
		FindStringCol("")
}

// FindServerSupportCNAME 查询服务是否支持CNAME
func (this *ServerDAO) FindServerSupportCNAME(tx *dbs.Tx, serverId int64) (bool, error) {
	supportCNAME, err := this.Query(tx).
		Pk(serverId).
		Result("supportCNAME").
		FindIntCol(0)
	if err != nil {
		return false, err
	}
	return supportCNAME == 1, nil
}

// FindStatelessServerDNS 查询服务的DNS相关信息，并且不关注状态
func (this *ServerDAO) FindStatelessServerDNS(tx *dbs.Tx, serverId int64) (*Server, error) {
	one, err := this.Query(tx).
		Pk(serverId).
		Result("id", "dnsName", "isOn", "state", "clusterId").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*Server), nil
}

// FindServerAdminIdAndUserId 获取当前服务的管理员ID和用户ID
func (this *ServerDAO) FindServerAdminIdAndUserId(tx *dbs.Tx, serverId int64) (adminId int64, userId int64, err error) {
	one, err := this.Query(tx).
		Pk(serverId).
		Result("adminId", "userId").
		Find()
	if err != nil {
		return 0, 0, err
	}
	if one == nil {
		return 0, 0, nil
	}
	return int64(one.(*Server).AdminId), int64(one.(*Server).UserId), nil
}

// FindServerUserId  查找服务的用户ID
func (this *ServerDAO) FindServerUserId(tx *dbs.Tx, serverId int64) (userId int64, err error) {
	one, _, err := this.Query(tx).
		Pk(serverId).
		Result("userId").
		FindOne()
	if err != nil || one == nil {
		return 0, err
	}
	return one.GetInt64("userId"), nil
}

// FindServerUserPlanId  查找服务的套餐ID
func (this *ServerDAO) FindServerUserPlanId(tx *dbs.Tx, serverId int64) (userPlanId int64, err error) {
	return this.Query(tx).
		Pk(serverId).
		Result("userPlanId").
		FindInt64Col(0)
}

// CheckUserServer 检查用户服务
func (this *ServerDAO) CheckUserServer(tx *dbs.Tx, userId int64, serverId int64) error {
	if serverId <= 0 || userId <= 0 {
		return ErrNotFound
	}
	ok, err := this.Query(tx).
		Pk(serverId).
		Attr("userId", userId).
		Exist()
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	return nil
}

// UpdateUserServersClusterId 设置一个用户下的所有服务的所属集群
func (this *ServerDAO) UpdateUserServersClusterId(tx *dbs.Tx, userId int64, oldClusterId, newClusterId int64) error {
	_, err := this.Query(tx).
		Attr("userId", userId).
		Set("clusterId", newClusterId).
		Update()
	if err != nil {
		return err
	}

	if oldClusterId > 0 {
		err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, oldClusterId, NodeTaskTypeConfigChanged)
		if err != nil {
			return err
		}
		err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, oldClusterId, NodeTaskTypeIPItemChanged)
		if err != nil {
			return err
		}
		err = dns.SharedDNSTaskDAO.CreateClusterTask(tx, oldClusterId, dns.DNSTaskTypeClusterChange)
		if err != nil {
			return err
		}
	}

	if newClusterId > 0 {
		err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, newClusterId, NodeTaskTypeConfigChanged)
		if err != nil {
			return err
		}
		err = SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, newClusterId, NodeTaskTypeIPItemChanged)
		if err != nil {
			return err
		}
		err = dns.SharedDNSTaskDAO.CreateClusterTask(tx, newClusterId, dns.DNSTaskTypeClusterChange)
		if err != nil {
			return err
		}
	}

	return err
}

// FindAllEnabledServersWithUserId 查找用户的所有的服务
func (this *ServerDAO) FindAllEnabledServersWithUserId(tx *dbs.Tx, userId int64) (result []*Server, err error) {
	_, err = this.Query(tx).
		State(ServerStateEnabled).
		Attr("userId", userId).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindEnabledServerIdWithWebId 根据WebId查找ServerId
func (this *ServerDAO) FindEnabledServerIdWithWebId(tx *dbs.Tx, webId int64) (serverId int64, err error) {
	if webId <= 0 {
		return 0, nil
	}
	return this.Query(tx).
		State(ServerStateEnabled).
		Attr("webId", webId).
		ResultPk().
		FindInt64Col(0)
}

// FindEnabledServerIdWithReverseProxyId 查找包含某个反向代理的Server
func (this *ServerDAO) FindEnabledServerIdWithReverseProxyId(tx *dbs.Tx, reverseProxyId int64) (serverId int64, err error) {
	return this.Query(tx).
		State(ServerStateEnabled).
		Where("JSON_CONTAINS(reverseProxy, :jsonQuery)").
		Param("jsonQuery", maps.Map{"reverseProxyId": reverseProxyId}.AsJSON()).
		ResultPk().
		FindInt64Col(0)
}

// CheckTCPPortIsUsing 检查TCP端口是否被使用
func (this *ServerDAO) CheckTCPPortIsUsing(tx *dbs.Tx, clusterId int64, port int, excludeServerId int64, excludeProtocol string) (bool, error) {
	query := this.Query(tx).
		Attr("clusterId", clusterId).
		State(ServerStateEnabled).
		Param("port", types.String(port))
	if excludeServerId <= 0 {
		query.Where("JSON_CONTAINS(tcpPorts, :port)")
	} else {
		query.Where("(id!=:serverId AND JSON_CONTAINS(tcpPorts, :port))")
		query.Param("serverId", excludeServerId)
	}
	return query.
		Exist()
}

// ExistServerNameInCluster 检查ServerName是否已存在
func (this *ServerDAO) ExistServerNameInCluster(tx *dbs.Tx, clusterId int64, serverName string, excludeServerId int64) (bool, error) {
	query := this.Query(tx).
		Attr("clusterId", clusterId).
		Where("(JSON_CONTAINS(serverNames, :jsonQuery1) OR JSON_CONTAINS(serverNames, :jsonQuery2))").
		Param("jsonQuery1", maps.Map{"name": serverName}.AsJSON()).
		Param("jsonQuery2", maps.Map{"subNames": serverName}.AsJSON())
	if excludeServerId > 0 {
		query.Neq("id", excludeServerId)
	}
	query.State(ServerStateEnabled)
	return query.Exist()
}

// GenDNSName 生成DNS Name
func (this *ServerDAO) GenDNSName(tx *dbs.Tx) (string, error) {
	for {
		dnsName := rands.HexString(8)
		exist, err := this.Query(tx).
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

// FindLatestServers 查询最近访问的服务
func (this *ServerDAO) FindLatestServers(tx *dbs.Tx, size int64) (result []*Server, err error) {
	itemTable := SharedLatestItemDAO.Table
	itemType := LatestItemTypeServer
	_, err = this.Query(tx).
		Result(this.Table+".id", this.Table+".name").
		Join(SharedLatestItemDAO, dbs.QueryJoinRight, this.Table+".id="+itemTable+".itemId AND "+itemTable+".itemType='"+itemType+"'").
		Asc("CEIL((UNIX_TIMESTAMP() - " + itemTable + ".updatedAt) / (7 * 86400))"). // 优先一个星期以内的
		Desc(itemTable + ".count").
		State(NodeClusterStateEnabled).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindNearbyServersInGroup 查找所属分组附近的服务
func (this *ServerDAO) FindNearbyServersInGroup(tx *dbs.Tx, groupId int64, serverId int64, size int64) (result []*Server, err error) {
	// 之前的
	ones, err := SharedServerDAO.Query(tx).
		Result("id", "name", "isOn").
		State(ServerStateEnabled).
		Where("JSON_CONTAINS(groupIds, :groupId)").
		Param("groupId", numberutils.FormatInt64(groupId)).
		Gte("id", serverId).
		Limit(size).
		AscPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	lists.Reverse(ones)
	for _, one := range ones {
		result = append(result, one.(*Server))
	}

	// 之后的
	ones, err = SharedServerDAO.Query(tx).
		Result("id", "name", "isOn").
		State(ServerStateEnabled).
		Where("JSON_CONTAINS(groupIds, :groupId)").
		Param("groupId", numberutils.FormatInt64(groupId)).
		Lt("id", serverId).
		Limit(size).
		DescPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		result = append(result, one.(*Server))
	}
	return
}

// FindNearbyServersInCluster 查找所属集群附近的服务
func (this *ServerDAO) FindNearbyServersInCluster(tx *dbs.Tx, clusterId int64, serverId int64, size int64) (result []*Server, err error) {
	// 之前的
	ones, err := SharedServerDAO.Query(tx).
		Result("id", "name", "isOn").
		State(ServerStateEnabled).
		Attr("clusterId", clusterId).
		Gte("id", serverId).
		Limit(size).
		AscPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	lists.Reverse(ones)
	for _, one := range ones {
		result = append(result, one.(*Server))
	}

	// 之后的
	ones, err = SharedServerDAO.Query(tx).
		Result("id", "name", "isOn").
		State(ServerStateEnabled).
		Attr("clusterId", clusterId).
		Lt("id", serverId).
		Limit(size).
		DescPk().
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		result = append(result, one.(*Server))
	}
	return
}

// FindFirstHTTPOrHTTPSPortWithClusterId 获取集群中第一个HTTP或者HTTPS端口
func (this *ServerDAO) FindFirstHTTPOrHTTPSPortWithClusterId(tx *dbs.Tx, clusterId int64) (int, error) {
	one, _, err := this.Query(tx).
		Result("JSON_EXTRACT(http, '$.listen[*].portRange') AS httpPort, JSON_EXTRACT(https, '$.listen[*].portRange') AS httpsPort").
		Attr("clusterId", clusterId).
		State(ServerStateEnabled).
		Attr("isOn", 1).
		Where("((JSON_CONTAINS(http, :queryJSON) AND JSON_EXTRACT(http, '$.listen[*].portRange') IS NOT NULL) OR (JSON_CONTAINS(https, :queryJSON) AND JSON_EXTRACT(https, '$.listen[*].portRange') IS NOT NULL))").
		Param("queryJSON", "{\"isOn\":true}").
		FindOne()
	if err != nil {
		return 0, err
	}
	httpPortString := one.GetString("httpPort")
	if len(httpPortString) > 0 {
		var ports = []string{}
		err = json.Unmarshal([]byte(httpPortString), &ports)
		if err != nil {
			return 0, err
		}
		if len(ports) > 0 {
			var port = ports[0]
			if strings.Contains(port, "-") { // IP范围
				return types.Int(port[:strings.Index(port, "-")]), nil
			}
			return types.Int(port), nil
		}
	}

	httpsPortString := one.GetString("httpsPort")
	if len(httpsPortString) > 0 {
		var ports = []string{}
		err = json.Unmarshal([]byte(httpsPortString), &ports)
		if err != nil {
			return 0, err
		}
		var port = ports[0]

		if strings.Contains(port, "-") { // IP范围
			return types.Int(port[:strings.Index(port, "-")]), nil
		}
		return types.Int(port), nil
	}

	return 0, nil
}

// NotifyServerPortsUpdate 通知服务端口变化
func (this *ServerDAO) NotifyServerPortsUpdate(tx *dbs.Tx, serverId int64) error {
	one, err := this.Query(tx).
		Pk(serverId).
		Result("tcp", "tls", "udp", "http", "https").
		Find()
	if err != nil {
		return err
	}
	if one == nil {
		return nil
	}
	var server = one.(*Server)

	// HTTP
	var tcpListens = []*serverconfigs.NetworkAddressConfig{}
	var udpListens = []*serverconfigs.NetworkAddressConfig{}
	if len(server.Http) > 0 && server.Http != "null" {
		httpConfig := &serverconfigs.HTTPProtocolConfig{}
		err := json.Unmarshal([]byte(server.Http), httpConfig)
		if err != nil {
			return err
		}
		tcpListens = append(tcpListens, httpConfig.Listen...)
	}

	// HTTPS
	if len(server.Https) > 0 && server.Https != "null" {
		httpsConfig := &serverconfigs.HTTPSProtocolConfig{}
		err := json.Unmarshal([]byte(server.Https), httpsConfig)
		if err != nil {
			return err
		}
		tcpListens = append(tcpListens, httpsConfig.Listen...)
	}

	// TCP
	if len(server.Tcp) > 0 && server.Tcp != "null" {
		tcpConfig := &serverconfigs.TCPProtocolConfig{}
		err := json.Unmarshal([]byte(server.Tcp), tcpConfig)
		if err != nil {
			return err
		}
		tcpListens = append(tcpListens, tcpConfig.Listen...)
	}

	// TLS
	if len(server.Tls) > 0 && server.Tls != "null" {
		tlsConfig := &serverconfigs.TLSProtocolConfig{}
		err := json.Unmarshal([]byte(server.Tls), tlsConfig)
		if err != nil {
			return err
		}
		tcpListens = append(tcpListens, tlsConfig.Listen...)
	}

	// UDP
	if len(server.Udp) > 0 && server.Udp != "null" {
		udpConfig := &serverconfigs.UDPProtocolConfig{}
		err := json.Unmarshal([]byte(server.Udp), udpConfig)
		if err != nil {
			return err
		}
		udpListens = append(udpListens, udpConfig.Listen...)
	}

	var tcpPorts = []int{}
	for _, listen := range tcpListens {
		_ = listen.Init()
		if listen.MinPort > 0 && listen.MaxPort > 0 && listen.MinPort <= listen.MaxPort {
			for i := listen.MinPort; i <= listen.MaxPort; i++ {
				if !lists.ContainsInt(tcpPorts, i) {
					tcpPorts = append(tcpPorts, i)
				}
			}
		}
	}

	tcpPortsJSON, err := json.Marshal(tcpPorts)
	if err != nil {
		return err
	}

	var udpPorts = []int{}
	for _, listen := range udpListens {
		_ = listen.Init()
		if listen.MinPort > 0 && listen.MaxPort > 0 && listen.MinPort <= listen.MaxPort {
			for i := listen.MinPort; i <= listen.MaxPort; i++ {
				if !lists.ContainsInt(udpPorts, i) {
					udpPorts = append(udpPorts, i)
				}
			}
		}
	}

	udpPortsJSON, err := json.Marshal(udpPorts)
	if err != nil {
		return err
	}

	return this.Query(tx).
		Pk(serverId).
		Set("tcpPorts", string(tcpPortsJSON)).
		Set("udpPorts", string(udpPortsJSON)).
		UpdateQuickly()
}

// FindServerTrafficLimitConfig 查找服务的流量限制
func (this *ServerDAO) FindServerTrafficLimitConfig(tx *dbs.Tx, serverId int64, cacheMap maps.Map) (*serverconfigs.TrafficLimitConfig, error) {
	if cacheMap == nil {
		cacheMap = maps.Map{}
	}
	var cacheKey = this.Table + ":FindServerTrafficLimitConfig:" + types.String(serverId)
	result, ok := cacheMap[cacheKey]
	if ok {
		return result.(*serverconfigs.TrafficLimitConfig), nil
	}

	serverOne, err := this.Query(tx).
		Pk(serverId).
		Result("trafficLimit").
		Find()
	if err != nil {
		return nil, err
	}

	var limit = &serverconfigs.TrafficLimitConfig{}
	if serverOne == nil {
		return limit, nil
	}

	var trafficLimit = serverOne.(*Server).TrafficLimit

	err = json.Unmarshal([]byte(trafficLimit), limit)
	if err != nil {
		return nil, err
	}

	cacheMap[cacheKey] = limit

	return limit, nil
}

// CalculateServerTrafficLimitConfig 计算服务的流量限制
// TODO 优化性能
func (this *ServerDAO) CalculateServerTrafficLimitConfig(tx *dbs.Tx, serverId int64, cacheMap maps.Map) (*serverconfigs.TrafficLimitConfig, error) {
	if cacheMap == nil {
		cacheMap = maps.Map{}
	}
	var cacheKey = this.Table + ":FindServerTrafficLimitConfig:" + types.String(serverId)
	result, ok := cacheMap[cacheKey]
	if ok {
		return result.(*serverconfigs.TrafficLimitConfig), nil
	}

	serverOne, err := this.Query(tx).
		Pk(serverId).
		Result("trafficLimit", "userPlanId").
		Find()
	if err != nil {
		return nil, err
	}

	var limitConfig = &serverconfigs.TrafficLimitConfig{}
	if serverOne == nil {
		return limitConfig, nil
	}

	var trafficLimit = serverOne.(*Server).TrafficLimit
	var userPlanId = int64(serverOne.(*Server).UserPlanId)

	if len(trafficLimit) == 0 {
		if userPlanId > 0 {
			userPlan, err := SharedUserPlanDAO.FindEnabledUserPlan(tx, userPlanId, cacheMap)
			if err != nil {
				return nil, err
			}
			if userPlan != nil {
				planLimit, err := SharedPlanDAO.FindEnabledPlanTrafficLimit(tx, int64(userPlan.PlanId), cacheMap)
				if err != nil {
					return nil, err
				}
				if planLimit != nil {
					return planLimit, nil
				}
			}
		}
		return limitConfig, nil
	}

	err = json.Unmarshal([]byte(trafficLimit), limitConfig)
	if err != nil {
		return nil, err
	}

	if !limitConfig.IsOn {
		if userPlanId > 0 {
			userPlan, err := SharedUserPlanDAO.FindEnabledUserPlan(tx, userPlanId, cacheMap)
			if err != nil {
				return nil, err
			}
			if userPlan != nil {
				planLimit, err := SharedPlanDAO.FindEnabledPlanTrafficLimit(tx, int64(userPlan.PlanId), cacheMap)
				if err != nil {
					return nil, err
				}
				if planLimit != nil {
					return planLimit, nil
				}
			}
		}
	}

	cacheMap[cacheKey] = limitConfig

	return limitConfig, nil
}

// UpdateServerTrafficLimitConfig 修改服务的流量限制
func (this *ServerDAO) UpdateServerTrafficLimitConfig(tx *dbs.Tx, serverId int64, trafficLimitConfig *serverconfigs.TrafficLimitConfig) error {
	if serverId <= 0 {
		return errors.New("invalid serverId")
	}
	limitJSON, err := json.Marshal(trafficLimitConfig)
	if err != nil {
		return err
	}

	err = this.Query(tx).
		Pk(serverId).
		Set("trafficLimit", limitJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}

	// 更新状态
	return this.UpdateServerTrafficLimitStatus(tx, trafficLimitConfig, serverId, true)
}

func (this *ServerDAO) UpdateServerTrafficLimitStatus(tx *dbs.Tx, trafficLimitConfig *serverconfigs.TrafficLimitConfig, serverId int64, isUpdatingConfig bool) error {
	if !trafficLimitConfig.IsOn {
		if isUpdatingConfig {
			return this.NotifyUpdate(tx, serverId)
		}
		return nil
	}

	serverOne, err := this.Query(tx).
		Pk(serverId).
		Result("trafficLimitStatus", "totalTraffic", "totalDailyTraffic", "totalMonthlyTraffic", "trafficDay", "trafficMonth").
		Find()
	if err != nil {
		return err
	}
	if serverOne == nil {
		return nil
	}

	var server = serverOne.(*Server)

	var oldStatus = &serverconfigs.TrafficLimitStatus{}
	if len(server.TrafficLimitStatus) > 0 {
		err = json.Unmarshal([]byte(server.TrafficLimitStatus), oldStatus)
		if err != nil {
			return err
		}

		// 如果已经达到限制了，而且还在有效期，那就没必要再更新
		if !isUpdatingConfig && oldStatus.IsValid() {
			return nil
		}
	}

	var untilDay = ""

	// daily
	if trafficLimitConfig.DailyBytes() > 0 {
		if server.TrafficDay == timeutil.Format("Ymd") && server.TotalDailyTraffic >= float64(trafficLimitConfig.DailyBytes())/1024/1024/1024 {
			untilDay = timeutil.Format("Ymd")
		}
	}

	// monthly
	if server.TrafficMonth == timeutil.Format("Ym") && trafficLimitConfig.MonthlyBytes() > 0 {
		if server.TotalMonthlyTraffic >= float64(trafficLimitConfig.MonthlyBytes())/1024/1024/1024 {
			untilDay = timeutil.Format("Ym32")
		}
	}

	// totally
	if trafficLimitConfig.TotalBytes() > 0 {
		if server.TotalTraffic >= float64(trafficLimitConfig.TotalBytes())/1024/1024/1024 {
			untilDay = "30000101"
		}
	}

	var isChanged = oldStatus.UntilDay != untilDay
	if isChanged {
		statusJSON, err := json.Marshal(&serverconfigs.TrafficLimitStatus{UntilDay: untilDay})
		if err != nil {
			return err
		}

		err = this.Query(tx).
			Pk(serverId).
			Set("trafficLimitStatus", statusJSON).
			UpdateQuickly()
		if err != nil {
			return err
		}
		return this.NotifyUpdate(tx, serverId)
	}

	if isUpdatingConfig {
		return this.NotifyUpdate(tx, serverId)
	}
	return nil
}

// IncreaseServerTotalTraffic 增加服务的总流量
func (this *ServerDAO) IncreaseServerTotalTraffic(tx *dbs.Tx, serverId int64, bytes int64) error {
	var gb = float64(bytes) / 1024 / 1024 / 1024
	var day = timeutil.Format("Ymd")
	var month = timeutil.Format("Ym")
	return this.Query(tx).
		Pk(serverId).
		Set("totalDailyTraffic", dbs.SQL("IF(trafficDay=:day, totalDailyTraffic, 0)+:trafficGB")).
		Set("totalMonthlyTraffic", dbs.SQL("IF(trafficMonth=:month, totalMonthlyTraffic, 0)+:trafficGB")).
		Set("totalTraffic", dbs.SQL("totalTraffic+:trafficGB")).
		Set("trafficDay", day).
		Set("trafficMonth", month).
		Param("day", day).
		Param("month", month).
		Param("trafficGB", gb).
		UpdateQuickly()

}

// ResetServerTotalTraffic 重置服务总流量
func (this *ServerDAO) ResetServerTotalTraffic(tx *dbs.Tx, serverId int64) error {
	return this.Query(tx).
		Pk(serverId).
		Set("totalDailyTraffic", 0).
		Set("totalMonthlyTraffic", 0).
		UpdateQuickly()
}

// FindEnabledServerIdWithUserPlanId 查找使用某个套餐的服务
func (this *ServerDAO) FindEnabledServerIdWithUserPlanId(tx *dbs.Tx, userPlanId int64) (int64, error) {
	return this.Query(tx).
		State(ServerStateEnabled).
		Attr("userPlanId", userPlanId).
		ResultPk().
		FindInt64Col(0)
}

// UpdateServersClusterIdWithPlanId 修改套餐所在集群
func (this *ServerDAO) UpdateServersClusterIdWithPlanId(tx *dbs.Tx, planId int64, clusterId int64) error {
	return this.Query(tx).
		Where("userPlanId IN (SELECT id FROM "+SharedUserPlanDAO.Table+" WHERE planId=:planId AND state=1)").
		Param("planId", planId).
		Set("clusterId", clusterId).
		UpdateQuickly()
}

// UpdateServerUserPlanId 设置服务所属套餐
func (this *ServerDAO) UpdateServerUserPlanId(tx *dbs.Tx, serverId int64, userPlanId int64) error {
	userPlan, err := SharedUserPlanDAO.FindEnabledUserPlan(tx, userPlanId, nil)
	if err != nil {
		return err
	}
	if userPlan == nil {
		return errors.New("can not find user plan with id '" + types.String(userPlanId) + "'")
	}

	plan, err := SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
	if err != nil {
		return err
	}
	if plan == nil {
		return errors.New("can not find plan with id '" + types.String(userPlan.PlanId) + "'")
	}

	err = this.Query(tx).
		Pk(serverId).
		Set("userPlanId", userPlanId).
		Set("clusterId", plan.ClusterId).
		UpdateQuickly()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, serverId)
}

// NotifyUpdate 同步集群
func (this *ServerDAO) NotifyUpdate(tx *dbs.Tx, serverId int64) error {
	// 创建任务
	clusterId, err := this.FindServerClusterId(tx, serverId)
	if err != nil {
		return err
	}
	if clusterId == 0 {
		return nil
	}
	return SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleNode, clusterId, NodeTaskTypeConfigChanged)
}

// NotifyDNSUpdate 通知DNS更新
func (this *ServerDAO) NotifyDNSUpdate(tx *dbs.Tx, serverId int64) error {
	clusterId, err := this.Query(tx).
		Pk(serverId).
		Result("clusterId").
		FindInt64Col(0) // 这里不需要加服务状态条件，因为我们即使删除也要删除对应的服务的DNS解析
	if err != nil {
		return err
	}
	if clusterId <= 0 {
		return nil
	}
	dnsInfo, err := SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId, nil)
	if err != nil {
		return err
	}
	if dnsInfo == nil {
		return nil
	}
	if len(dnsInfo.DnsName) == 0 || dnsInfo.DnsDomainId <= 0 {
		return nil
	}
	return dns.SharedDNSTaskDAO.CreateServerTask(tx, serverId, dns.DNSTaskTypeServerChange)
}
