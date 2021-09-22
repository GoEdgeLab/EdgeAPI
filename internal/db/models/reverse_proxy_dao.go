package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

const (
	ReverseProxyStateEnabled  = 1 // 已启用
	ReverseProxyStateDisabled = 0 // 已禁用
)

type ReverseProxyDAO dbs.DAO

func NewReverseProxyDAO() *ReverseProxyDAO {
	return dbs.NewDAO(&ReverseProxyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeReverseProxies",
			Model:  new(ReverseProxy),
			PkName: "id",
		},
	}).(*ReverseProxyDAO)
}

var SharedReverseProxyDAO *ReverseProxyDAO

func init() {
	dbs.OnReady(func() {
		SharedReverseProxyDAO = NewReverseProxyDAO()
	})
}

// Init 初始化
func (this *ReverseProxyDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableReverseProxy 启用条目
func (this *ReverseProxyDAO) EnableReverseProxy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ReverseProxyStateEnabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, id)
}

// DisableReverseProxy 禁用条目
func (this *ReverseProxyDAO) DisableReverseProxy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ReverseProxyStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, id)
}

// FindEnabledReverseProxy 查找启用中的条目
func (this *ReverseProxyDAO) FindEnabledReverseProxy(tx *dbs.Tx, id int64) (*ReverseProxy, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ReverseProxyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ReverseProxy), err
}

// ComposeReverseProxyConfig 根据ID组合配置
func (this *ReverseProxyDAO) ComposeReverseProxyConfig(tx *dbs.Tx, reverseProxyId int64, cacheMap maps.Map) (*serverconfigs.ReverseProxyConfig, error) {
	if cacheMap == nil {
		cacheMap = maps.Map{}
	}
	var cacheKey = this.Table + ":config:" + types.String(reverseProxyId)
	var cache = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*serverconfigs.ReverseProxyConfig), nil
	}

	reverseProxy, err := this.FindEnabledReverseProxy(tx, reverseProxyId)
	if err != nil {
		return nil, err
	}
	if reverseProxy == nil {
		return nil, nil
	}

	config := &serverconfigs.ReverseProxyConfig{}
	config.Id = int64(reverseProxy.Id)
	config.IsOn = reverseProxy.IsOn == 1
	config.RequestHostType = types.Int8(reverseProxy.RequestHostType)
	config.RequestHost = reverseProxy.RequestHost
	config.RequestURI = reverseProxy.RequestURI
	config.StripPrefix = reverseProxy.StripPrefix
	config.AutoFlush = reverseProxy.AutoFlush == 1

	schedulingConfig := &serverconfigs.SchedulingConfig{}
	if len(reverseProxy.Scheduling) > 0 && reverseProxy.Scheduling != "null" {
		err = json.Unmarshal([]byte(reverseProxy.Scheduling), schedulingConfig)
		if err != nil {
			return nil, err
		}
		config.Scheduling = schedulingConfig
	}
	if len(reverseProxy.PrimaryOrigins) > 0 && reverseProxy.PrimaryOrigins != "null" {
		originRefs := []*serverconfigs.OriginRef{}
		err = json.Unmarshal([]byte(reverseProxy.PrimaryOrigins), &originRefs)
		if err != nil {
			return nil, err
		}
		for _, ref := range originRefs {
			originConfig, err := SharedOriginDAO.ComposeOriginConfig(tx, ref.OriginId, cacheMap)
			if err != nil {
				return nil, err
			}
			if originConfig != nil {
				config.AddPrimaryOrigin(originConfig)
			}
		}
	}

	if len(reverseProxy.BackupOrigins) > 0 && reverseProxy.BackupOrigins != "null" {
		originRefs := []*serverconfigs.OriginRef{}
		err = json.Unmarshal([]byte(reverseProxy.BackupOrigins), &originRefs)
		if err != nil {
			return nil, err
		}
		for _, originConfig := range originRefs {
			originConfig, err := SharedOriginDAO.ComposeOriginConfig(tx, originConfig.OriginId, cacheMap)
			if err != nil {
				return nil, err
			}
			if originConfig != nil {
				config.AddBackupOrigin(originConfig)
			}
		}
	}

	// add headers
	if IsNotNull(reverseProxy.AddHeaders) {
		addHeaders := []string{}
		err = json.Unmarshal([]byte(reverseProxy.AddHeaders), &addHeaders)
		if err != nil {
			return nil, err
		}
		config.AddHeaders = addHeaders
	}

	// 源站相关默认设置
	config.MaxConns = int(reverseProxy.MaxConns)
	config.MaxIdleConns = int(reverseProxy.MaxIdleConns)

	if IsNotNull(reverseProxy.ConnTimeout) {
		connTimeout := &shared.TimeDuration{}
		err = json.Unmarshal([]byte(reverseProxy.ConnTimeout), &connTimeout)
		if err != nil {
			return nil, err
		}
		config.ConnTimeout = connTimeout
	}

	if IsNotNull(reverseProxy.ReadTimeout) {
		readTimeout := &shared.TimeDuration{}
		err = json.Unmarshal([]byte(reverseProxy.ReadTimeout), &readTimeout)
		if err != nil {
			return nil, err
		}
		config.ReadTimeout = readTimeout
	}

	if IsNotNull(reverseProxy.IdleTimeout) {
		idleTimeout := &shared.TimeDuration{}
		err = json.Unmarshal([]byte(reverseProxy.IdleTimeout), &idleTimeout)
		if err != nil {
			return nil, err
		}
		config.IdleTimeout = idleTimeout
	}

	cacheMap[cacheKey] = config

	return config, nil
}

// CreateReverseProxy 创建反向代理
func (this *ReverseProxyDAO) CreateReverseProxy(tx *dbs.Tx, adminId int64, userId int64, schedulingJSON []byte, primaryOriginsJSON []byte, backupOriginsJSON []byte) (int64, error) {
	op := NewReverseProxyOperator()
	op.IsOn = true
	op.State = ReverseProxyStateEnabled
	op.AdminId = adminId
	op.UserId = userId
	op.RequestHostType = serverconfigs.RequestHostTypeProxyServer

	defaultHeaders := []string{"X-Real-IP", "X-Forwarded-For", "X-Forwarded-By", "X-Forwarded-Host", "X-Forwarded-Proto"}
	defaultHeadersJSON, err := json.Marshal(defaultHeaders)
	if err != nil {
		return 0, err
	}
	op.AddHeaders = defaultHeadersJSON

	if len(schedulingJSON) > 0 {
		op.Scheduling = string(schedulingJSON)
	}
	if len(primaryOriginsJSON) > 0 {
		op.PrimaryOrigins = string(primaryOriginsJSON)
	}
	if len(backupOriginsJSON) > 0 {
		op.BackupOrigins = string(backupOriginsJSON)
	}
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// UpdateReverseProxyScheduling 修改反向代理调度算法
func (this *ReverseProxyDAO) UpdateReverseProxyScheduling(tx *dbs.Tx, reverseProxyId int64, schedulingJSON []byte) error {
	if reverseProxyId <= 0 {
		return errors.New("invalid reverseProxyId")
	}
	op := NewReverseProxyOperator()
	op.Id = reverseProxyId
	if len(schedulingJSON) > 0 {
		op.Scheduling = string(schedulingJSON)
	} else {
		op.Scheduling = "null"
	}
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, reverseProxyId)
}

// UpdateReverseProxyPrimaryOrigins 修改主要源站
func (this *ReverseProxyDAO) UpdateReverseProxyPrimaryOrigins(tx *dbs.Tx, reverseProxyId int64, origins []byte) error {
	if reverseProxyId <= 0 {
		return errors.New("invalid reverseProxyId")
	}
	op := NewReverseProxyOperator()
	op.Id = reverseProxyId
	if len(origins) > 0 {
		op.PrimaryOrigins = origins
	} else {
		op.PrimaryOrigins = "[]"
	}
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, reverseProxyId)
}

// UpdateReverseProxyBackupOrigins 修改备用源站
func (this *ReverseProxyDAO) UpdateReverseProxyBackupOrigins(tx *dbs.Tx, reverseProxyId int64, origins []byte) error {
	if reverseProxyId <= 0 {
		return errors.New("invalid reverseProxyId")
	}
	op := NewReverseProxyOperator()
	op.Id = reverseProxyId
	if len(origins) > 0 {
		op.BackupOrigins = origins
	} else {
		op.BackupOrigins = "[]"
	}
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, reverseProxyId)
}

// UpdateReverseProxy 修改是否启用
func (this *ReverseProxyDAO) UpdateReverseProxy(tx *dbs.Tx, reverseProxyId int64, requestHostType int8, requestHost string, requestURI string, stripPrefix string, autoFlush bool, addHeaders []string, connTimeout *shared.TimeDuration, readTimeout *shared.TimeDuration, idleTimeout *shared.TimeDuration, maxConns int32, maxIdleConns int32) error {
	if reverseProxyId <= 0 {
		return errors.New("invalid reverseProxyId")
	}

	op := NewReverseProxyOperator()
	op.Id = reverseProxyId

	if requestHostType < 0 {
		requestHostType = 0
	}
	op.RequestHostType = requestHostType

	op.RequestHost = requestHost
	op.RequestURI = requestURI
	op.StripPrefix = stripPrefix
	op.AutoFlush = autoFlush

	if len(addHeaders) == 0 {
		addHeaders = []string{}
	}
	addHeadersJSON, err := json.Marshal(addHeaders)
	if err != nil {
		return err
	}
	op.AddHeaders = addHeadersJSON

	if connTimeout != nil {
		connTimeoutJSON, err := connTimeout.AsJSON()
		if err != nil {
			return err
		}
		op.ConnTimeout = connTimeoutJSON
	}
	if readTimeout != nil {
		readTimeoutJSON, err := readTimeout.AsJSON()
		if err != nil {
			return err
		}
		op.ReadTimeout = readTimeoutJSON
	}
	if idleTimeout != nil {
		idleTimeoutJSON, err := idleTimeout.AsJSON()
		if err != nil {
			return err
		}
		op.IdleTimeout = idleTimeoutJSON
	}
	if maxConns >= 0 {
		op.MaxConns = maxConns
	} else {
		op.MaxConns = 0
	}
	if maxIdleConns >= 0 {
		op.MaxIdleConns = maxIdleConns
	} else {
		op.MaxIdleConns = 0
	}

	err = this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, reverseProxyId)
}

// FindReverseProxyContainsOriginId 查找包含某个源站的反向代理ID
func (this *ReverseProxyDAO) FindReverseProxyContainsOriginId(tx *dbs.Tx, originId int64) (int64, error) {
	return this.Query(tx).
		ResultPk().
		Where("(JSON_CONTAINS(primaryOrigins, :jsonQuery) OR JSON_CONTAINS(backupOrigins, :jsonQuery))").
		Param("jsonQuery", maps.Map{
			"originId": originId,
		}.AsJSON()).
		FindInt64Col(0)
}

// CheckUserReverseProxy 检查用户权限
func (this *ReverseProxyDAO) CheckUserReverseProxy(tx *dbs.Tx, userId int64, reverseProxyId int64) error {
	exists, err := this.Query(tx).
		Pk(reverseProxyId).
		Attr("userId", userId).
		Exist()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// 检查server是否为用户的
	serverId, err := SharedServerDAO.FindEnabledServerIdWithReverseProxyId(tx, reverseProxyId)
	if err != nil {
		return err
	}
	if serverId == 0 {
		return ErrNotFound
	}
	return SharedServerDAO.CheckUserServer(tx, userId, serverId)
}

// NotifyUpdate 通知更新
func (this *ReverseProxyDAO) NotifyUpdate(tx *dbs.Tx, reverseProxyId int64) error {
	serverId, err := SharedServerDAO.FindEnabledServerIdWithReverseProxyId(tx, reverseProxyId)
	if err != nil {
		return err
	}
	if serverId > 0 {
		return SharedServerDAO.NotifyUpdate(tx, serverId)
	}

	// locations
	locationId, err := SharedHTTPLocationDAO.FindEnabledLocationIdWithReverseProxyId(tx, reverseProxyId)
	if err != nil {
		return err
	}
	if locationId > 0 {
		return SharedHTTPLocationDAO.NotifyUpdate(tx, locationId)
	}

	// group
	groupId, err := SharedServerGroupDAO.FindEnabledGroupIdWithReverseProxyId(tx, reverseProxyId)
	if err != nil {
		return err
	}
	if groupId > 0 {
		return SharedServerGroupDAO.NotifyUpdate(tx, groupId)
	}

	return nil
}
