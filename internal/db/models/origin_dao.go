package models

import (
	"encoding/json"
	"errors"
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

var SharedOriginDAO = NewOriginDAO()

// 初始化
func (this *OriginDAO) Init() {
	this.DAOObject.Init()
	this.DAOObject.OnUpdate(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
	this.DAOObject.OnInsert(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
	this.DAOObject.OnDelete(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
}

// 启用条目
func (this *OriginDAO) EnableOrigin(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", OriginStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *OriginDAO) DisableOrigin(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", OriginStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *OriginDAO) FindEnabledOrigin(id int64) (*Origin, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", OriginStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Origin), err
}

// 根据主键查找名称
func (this *OriginDAO) FindOriginName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建源站
func (this *OriginDAO) CreateOrigin(name string, addrJSON string, description string) (originId int64, err error) {
	op := NewOriginOperator()
	op.IsOn = true
	op.Name = name
	op.Addr = addrJSON
	op.Description = description
	op.State = OriginStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return
	}
	return types.Int64(op.Id), nil
}

// 修改源站
func (this *OriginDAO) UpdateOrigin(originId int64, name string, addrJSON string, description string) error {
	if originId <= 0 {
		return errors.New("invalid originId")
	}
	op := NewOriginOperator()
	op.Id = originId
	op.Name = name
	op.Addr = addrJSON
	op.Description = description
	op.Version = dbs.SQL("version+1")
	_, err := this.Save(op)
	return err
}

// 将源站信息转换为配置
func (this *OriginDAO) ComposeOriginConfig(originId int64) (*serverconfigs.OriginConfig, error) {
	origin, err := this.FindEnabledOrigin(originId)
	if err != nil {
		return nil, err
	}
	if origin == nil {
		return nil, errors.New("not found")
	}

	config := &serverconfigs.OriginConfig{
		Id:           int64(origin.Id),
		IsOn:         origin.IsOn == 1,
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
	}

	if len(origin.Addr) > 0 && origin.Addr != "null" {
		addr := &serverconfigs.NetworkAddressConfig{}
		err = json.Unmarshal([]byte(origin.Addr), addr)
		if err != nil {
			return nil, err
		}
		config.Addr = addr
	}

	if len(origin.ConnTimeout) > 0 && origin.ConnTimeout != "null" {
		connTimeout := &shared.TimeDuration{}
		err = json.Unmarshal([]byte(origin.ConnTimeout), &connTimeout)
		if err != nil {
			return nil, err
		}
		config.ConnTimeout = connTimeout
	}

	if len(origin.ReadTimeout) > 0 && origin.ReadTimeout != "null" {
		readTimeout := &shared.TimeDuration{}
		err = json.Unmarshal([]byte(origin.ReadTimeout), &readTimeout)
		if err != nil {
			return nil, err
		}
		config.ReadTimeout = readTimeout
	}

	if len(origin.IdleTimeout) > 0 && origin.IdleTimeout != "null" {
		idleTimeout := &shared.TimeDuration{}
		err = json.Unmarshal([]byte(origin.IdleTimeout), &idleTimeout)
		if err != nil {
			return nil, err
		}
		config.IdleTimeout = idleTimeout
	}

	// headers
	if IsNotNull(origin.HttpRequestHeader) {
		ref := &shared.HTTPHeaderPolicyRef{}
		err = json.Unmarshal([]byte(origin.HttpRequestHeader), ref)
		if err != nil {
			return nil, err
		}
		config.RequestHeaderPolicyRef = ref

		if ref.HeaderPolicyId > 0 {
			headerPolicy, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(ref.HeaderPolicyId)
			if err != nil {
				return nil, err
			}
			if headerPolicy != nil {
				config.RequestHeaderPolicy = headerPolicy
			}
		}
	}

	if IsNotNull(origin.HttpResponseHeader) {
		ref := &shared.HTTPHeaderPolicyRef{}
		err = json.Unmarshal([]byte(origin.HttpResponseHeader), ref)
		if err != nil {
			return nil, err
		}
		config.ResponseHeaderPolicyRef = ref

		if ref.HeaderPolicyId > 0 {
			headerPolicy, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(ref.HeaderPolicyId)
			if err != nil {
				return nil, err
			}
			if headerPolicy != nil {
				config.ResponseHeaderPolicy = headerPolicy
			}
		}
	}

	if len(origin.HealthCheck) > 0 && origin.HealthCheck != "null" {
		healthCheck := &serverconfigs.HealthCheckConfig{}
		err = json.Unmarshal([]byte(origin.HealthCheck), healthCheck)
		if err != nil {
			return nil, err
		}
		config.HealthCheck = healthCheck
	}

	if len(origin.Cert) > 0 && origin.Cert != "null" {
		cert := &sslconfigs.SSLCertConfig{}
		err = json.Unmarshal([]byte(origin.Cert), cert)
		if err != nil {
			return nil, err
		}
		config.Cert = cert
	}

	if len(origin.Ftp) > 0 && origin.Ftp != "null" {
		ftp := &serverconfigs.OriginFTPConfig{}
		err = json.Unmarshal([]byte(origin.Ftp), ftp)
		if err != nil {
			return nil, err
		}
		config.FTP = ftp
	}

	return config, nil
}
