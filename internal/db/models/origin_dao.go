package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ossconfigs"
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
	addrJSON []byte,
	ossConfig *ossconfigs.OSSConfig,
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
	followPort bool,
	http2Enabled bool) (originId int64, err error) {
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

	if len(addrJSON) > 0 {
		op.Addr = addrJSON
	}

	if ossConfig != nil {
		ossConfigJSON, err := json.Marshal(ossConfig)
		if err != nil {
			return 0, err
		}
		op.Oss = ossConfigJSON
	}

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
	op.Http2Enabled = http2Enabled

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
	addrJSON []byte,
	ossConfig *ossconfigs.OSSConfig,
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
	followPort bool,
	http2Enabled bool) error {
	if originId <= 0 {
		return errors.New("invalid originId")
	}
	var op = NewOriginOperator()
	op.Id = originId
	op.Name = name

	op.Addr = addrJSON

	if ossConfig != nil {
		ossConfigJSON, err := json.Marshal(ossConfig)
		if err != nil {
			return err
		}
		op.Oss = ossConfigJSON
	} else {
		op.Oss = dbs.SQL("NULL")
	}

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
	op.Http2Enabled = http2Enabled

	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, originId)
}

// UpdateOriginIsOn 修改源站是否启用
func (this *OriginDAO) UpdateOriginIsOn(tx *dbs.Tx, originId int64, isOn bool) error {
	err := this.Query(tx).
		Pk(originId).
		Set("isOn", isOn).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, originId)
}

// CloneOrigin 复制源站
func (this *OriginDAO) CloneOrigin(tx *dbs.Tx, fromOriginId int64) (newOriginId int64, err error) {
	if fromOriginId <= 0 {
		return
	}
	originOne, err := this.Find(tx, fromOriginId)
	if err != nil || originOne == nil {
		return
	}
	var origin = originOne.(*Origin)
	var op = NewOriginOperator()
	op.IsOn = origin.IsOn
	op.Name = origin.Name
	op.Version = origin.Version
	if IsNotNull(origin.Addr) {
		op.Addr = origin.Addr
	}
	op.Description = origin.Description
	op.Code = origin.Code
	op.Weight = origin.Weight
	if IsNotNull(origin.ConnTimeout) {
		op.ConnTimeout = origin.ConnTimeout
	}
	if IsNotNull(origin.ReadTimeout) {
		op.ReadTimeout = origin.ReadTimeout
	}
	if IsNotNull(origin.IdleTimeout) {
		op.IdleTimeout = origin.IdleTimeout
	}
	op.MaxFails = origin.MaxFails
	op.MaxConns = origin.MaxConns
	op.MaxIdleConns = origin.MaxIdleConns
	op.HttpRequestURI = origin.HttpRequestURI
	if IsNotNull(origin.HttpRequestHeader) {
		op.HttpRequestHeader = origin.HttpRequestHeader
	}
	if IsNotNull(origin.HttpResponseHeader) {
		op.HttpResponseHeader = origin.HttpResponseHeader
	}
	op.Host = origin.Host
	if IsNotNull(origin.HealthCheck) {
		op.HealthCheck = origin.HealthCheck
	}
	if IsNotNull(origin.Cert) {
		// TODO 需要Clone证书
		op.Cert = origin.Cert
	}
	if IsNotNull(origin.Ftp) {
		op.Ftp = origin.Ftp
	}
	if IsNotNull(origin.Domains) {
		op.Domains = origin.Domains
	}
	op.FollowPort = origin.FollowPort
	op.Http2Enabled = origin.Http2Enabled
	op.State = origin.State
	return this.SaveInt64(tx, op)
}

// ComposeOriginConfig 将源站信息转换为配置
func (this *OriginDAO) ComposeOriginConfig(tx *dbs.Tx, originId int64, dataMap *shared.DataMap, cacheMap *utils.CacheMap) (*serverconfigs.OriginConfig, error) {
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
		HTTP2Enabled: origin.Http2Enabled,
	}

	// addr
	var isOSS = false
	if IsNotNull(origin.Addr) {
		var addr = &serverconfigs.NetworkAddressConfig{}
		err = json.Unmarshal(origin.Addr, addr)
		if err != nil {
			return nil, err
		}
		config.Addr = addr
		isOSS = ossconfigs.IsOSSProtocol(string(addr.Protocol))
	}

	// oss
	if isOSS && IsNotNull(origin.Oss) {
		var ossConfig = ossconfigs.NewOSSConfig()
		err = json.Unmarshal(origin.Oss, ossConfig)
		if err != nil {
			return nil, err
		}
		config.OSS = ossConfig
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
			certConfig, err := SharedSSLCertDAO.ComposeCertConfig(tx, ref.CertId, false, dataMap, cacheMap)
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

// ExistsOrigin 检查源站是否存在
func (this *OriginDAO) ExistsOrigin(tx *dbs.Tx, originId int64) (bool, error) {
	if originId <= 0 {
		return false, nil
	}
	return this.Query(tx).
		Pk(originId).
		State(OriginStateEnabled).
		Exist()
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
