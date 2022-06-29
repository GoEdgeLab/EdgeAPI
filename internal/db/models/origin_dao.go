package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	OriginStateEnabled  = 1 // 已启用
	OriginStateDisabled = 0 // 已禁用
)

type OriginDAO dbs.DAO

func NewOriginDAO() *OriginDAO {
	return dbs.NewDAO(&OriginDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeOrigins",
			Model:  new(Origin),
			PkName: "id",
		},
	}).(*OriginDAO)
}

var SharedOriginDAO *OriginDAO

func init() {
	dbs.OnReady(func() {
		SharedOriginDAO = NewOriginDAO()
	})
}

// Init 初始化
func (this *OriginDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableOrigin 启用条目
func (this *OriginDAO) EnableOrigin(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", OriginStateEnabled).
		Update()
	return err
}

// DisableOrigin 禁用条目
func (this *OriginDAO) DisableOrigin(tx *dbs.Tx, originId int64) error {
	_, err := this.Query(tx).
		Pk(originId).
		Set("state", OriginStateDisabled).
		Update()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, originId)
}

// FindEnabledOrigin 查找启用中的条目
func (this *OriginDAO) FindEnabledOrigin(tx *dbs.Tx, id int64) (*Origin, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", OriginStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Origin), err
}

// FindOriginName 根据主键查找名称
func (this *OriginDAO) FindOriginName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateOrigin 创建源站
func (this *OriginDAO) CreateOrigin(tx *dbs.Tx,
	adminId int64,
	userId int64,
	name string,
	addrJSON string,
	description string,
	weight int32, isOn bool,
	connTimeout *shared.TimeDuration,
	readTimeout *shared.TimeDuration,
	idleTimeout *shared.TimeDuration,
	maxConns int32,
	maxIdleConns int32,
	certRef *sslconfigs.SSLCertRef,
	domains []string,
	host string,
	followPort bool) (originId int64, err error) {
	var op = NewOriginOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.IsOn = isOn
	op.Name = name

	if connTimeout != nil {
		connTimeoutJSON, err := connTimeout.AsJSON()
		if err != nil {
			return 0, err
		}
		op.ConnTimeout = connTimeoutJSON
	}
	if readTimeout != nil {
		readTimeoutJSON, err := readTimeout.AsJSON()
		if err != nil {
			return 0, err
		}
		op.ReadTimeout = readTimeoutJSON
	}
	if idleTimeout != nil {
		idleTimeoutJSON, err := idleTimeout.AsJSON()
		if err != nil {
			return 0, err
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

	op.Addr = addrJSON
	op.Description = description
	if weight < 0 {
		weight = 0
	}
	op.Weight = weight

	// cert
	if certRef != nil {
		certRefJSON, err := json.Marshal(certRef)
		if err != nil {
			return 0, err
		}
		op.Cert = certRefJSON
	}

	if len(domains) > 0 {
		domainsJSON, err := json.Marshal(domains)
		if err != nil {
			return 0, err
		}
		op.Domains = domainsJSON
	} else {
		op.Domains = "[]"
	}

	op.Host = host
	op.FollowPort = followPort

	op.State = OriginStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return
	}
	return types.Int64(op.Id), nil
}

// UpdateOrigin 修改源站
func (this *OriginDAO) UpdateOrigin(tx *dbs.Tx,
	originId int64,
	name string,
	addrJSON string,
	description string,
	weight int32,
	isOn bool,
	connTimeout *shared.TimeDuration,
	readTimeout *shared.TimeDuration,
	idleTimeout *shared.TimeDuration,
	maxConns int32,
	maxIdleConns int32,
	certRef *sslconfigs.SSLCertRef,
	domains []string,
	host string,
	followPort bool) error {
	if originId <= 0 {
		return errors.New("invalid originId")
	}
	var op = NewOriginOperator()
	op.Id = originId
	op.Name = name
	op.Addr = addrJSON
	op.Description = description
	if weight < 0 {
		weight = 0
	}
	op.Weight = weight

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

	op.IsOn = isOn
	op.Version = dbs.SQL("version+1")

	// cert
	if certRef != nil {
		certRefJSON, err := json.Marshal(certRef)
		if err != nil {
			return err
		}
		op.Cert = certRefJSON
	} else {
		op.Cert = dbs.SQL("NULL")
	}

	if len(domains) > 0 {
		domainsJSON, err := json.Marshal(domains)
		if err != nil {
			return err
		}
		op.Domains = domainsJSON
	} else {
		op.Domains = "[]"
	}

	op.Host = host
	op.FollowPort = followPort

	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, originId)
}

// ComposeOriginConfig 将源站信息转换为配置
func (this *OriginDAO) ComposeOriginConfig(tx *dbs.Tx, originId int64, cacheMap *utils.CacheMap) (*serverconfigs.OriginConfig, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":config:" + types.String(originId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*serverconfigs.OriginConfig), nil
	}

	origin, err := this.FindEnabledOrigin(tx, originId)
	if err != nil {
		return nil, err
	}
	if origin == nil {
		return nil, nil
	}

	var config = &serverconfigs.OriginConfig{
		Id:           int64(origin.Id),
		IsOn:         origin.IsOn,
		Version:      int(origin.Version),
		Name:         origin.Name,
		Description:  origin.Description,
		Code:         origin.Code,
		Weight:       uint(origin.Weight),
		MaxFails:     int(origin.MaxFails),
		MaxConns:     int(origin.MaxConns),
		MaxIdleConns: int(origin.MaxIdleConns),
		RequestURI:   origin.HttpRequestURI,
		RequestHost:  origin.Host,
		Domains:      origin.DecodeDomains(),
		FollowPort:   origin.FollowPort,
	}

	if IsNotNull(origin.Addr) {
		var addr = &serverconfigs.NetworkAddressConfig{}
		err = json.Unmarshal(origin.Addr, addr)
		if err != nil {
			return nil, err
		}
		config.Addr = addr
	}

	if IsNotNull(origin.ConnTimeout) {
		var connTimeout = &shared.TimeDuration{}
		err = json.Unmarshal(origin.ConnTimeout, &connTimeout)
		if err != nil {
			return nil, err
		}
		config.ConnTimeout = connTimeout
	}

	if IsNotNull(origin.ReadTimeout) {
		var readTimeout = &shared.TimeDuration{}
		err = json.Unmarshal(origin.ReadTimeout, &readTimeout)
		if err != nil {
			return nil, err
		}
		config.ReadTimeout = readTimeout
	}

	if IsNotNull(origin.IdleTimeout) {
		var idleTimeout = &shared.TimeDuration{}
		err = json.Unmarshal(origin.IdleTimeout, &idleTimeout)
		if err != nil {
			return nil, err
		}
		config.IdleTimeout = idleTimeout
	}

	// headers
	if IsNotNull(origin.HttpRequestHeader) {
		ref := &shared.HTTPHeaderPolicyRef{}
		err = json.Unmarshal(origin.HttpRequestHeader, ref)
		if err != nil {
			return nil, err
		}
		config.RequestHeaderPolicyRef = ref

		if ref.HeaderPolicyId > 0 {
			headerPolicy, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(tx, ref.HeaderPolicyId)
			if err != nil {
				return nil, err
			}
			if headerPolicy != nil {
				config.RequestHeaderPolicy = headerPolicy
			}
		}
	}

	if IsNotNull(origin.HttpResponseHeader) {
		var ref = &shared.HTTPHeaderPolicyRef{}
		err = json.Unmarshal(origin.HttpResponseHeader, ref)
		if err != nil {
			return nil, err
		}
		config.ResponseHeaderPolicyRef = ref

		if ref.HeaderPolicyId > 0 {
			headerPolicy, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(tx, ref.HeaderPolicyId)
			if err != nil {
				return nil, err
			}
			if headerPolicy != nil {
				config.ResponseHeaderPolicy = headerPolicy
			}
		}
	}

	if IsNotNull(origin.HealthCheck) {
		var healthCheck = &serverconfigs.HealthCheckConfig{}
		err = json.Unmarshal(origin.HealthCheck, healthCheck)
		if err != nil {
			return nil, err
		}
		config.HealthCheck = healthCheck
	}

	if IsNotNull(origin.Cert) {
		var ref = &sslconfigs.SSLCertRef{}
		err = json.Unmarshal(origin.Cert, ref)
		if err != nil {
			return nil, err
		}
		config.CertRef = ref
		if ref.CertId > 0 {
			certConfig, err := SharedSSLCertDAO.ComposeCertConfig(tx, ref.CertId, cacheMap)
			if err != nil {
				return nil, err
			}
			config.Cert = certConfig
		}
	}

	if IsNotNull(origin.Ftp) {
		// TODO
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// CheckUserOrigin 检查源站权限
func (this *OriginDAO) CheckUserOrigin(tx *dbs.Tx, userId int64, originId int64) error {
	reverseProxyId, err := SharedReverseProxyDAO.FindReverseProxyContainsOriginId(tx, originId)
	if err != nil {
		return err
	}
	if reverseProxyId == 0 {
		// 这里我们不允许源站没有被使用
		return ErrNotFound
	}
	return SharedReverseProxyDAO.CheckUserReverseProxy(tx, userId, reverseProxyId)
}

// NotifyUpdate 通知更新
func (this *OriginDAO) NotifyUpdate(tx *dbs.Tx, originId int64) error {
	reverseProxyId, err := SharedReverseProxyDAO.FindReverseProxyContainsOriginId(tx, originId)
	if err != nil {
		return err
	}
	if reverseProxyId > 0 {
		return SharedReverseProxyDAO.NotifyUpdate(tx, reverseProxyId)
	}
	return nil
}
