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
	OriginServerStateEnabled  = 1 // 已启用
	OriginServerStateDisabled = 0 // 已禁用
)

type OriginServerDAO dbs.DAO

func NewOriginServerDAO() *OriginServerDAO {
	return dbs.NewDAO(&OriginServerDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeOriginServers",
			Model:  new(OriginServer),
			PkName: "id",
		},
	}).(*OriginServerDAO)
}

var SharedOriginServerDAO = NewOriginServerDAO()

// 启用条目
func (this *OriginServerDAO) EnableOriginServer(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", OriginServerStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *OriginServerDAO) DisableOriginServer(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", OriginServerStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *OriginServerDAO) FindEnabledOriginServer(id int64) (*OriginServer, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", OriginServerStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*OriginServer), err
}

// 根据主键查找名称
func (this *OriginServerDAO) FindOriginServerName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建源站
func (this *OriginServerDAO) CreateOriginServer(name string, addrJSON string, description string) (originId int64, err error) {
	op := NewOriginServerOperator()
	op.IsOn = true
	op.Name = name
	op.Addr = addrJSON
	op.Description = description
	op.State = OriginServerStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return
	}
	return types.Int64(op.Id), nil
}

// 修改源站
func (this *OriginServerDAO) UpdateOriginServer(originId int64, name string, addrJSON string, description string) error {
	if originId <= 0 {
		return errors.New("invalid originId")
	}
	op := NewOriginServerOperator()
	op.Id = originId
	op.Name = name
	op.Addr = addrJSON
	op.Description = description
	op.Version = dbs.SQL("version+1")
	_, err := this.Save(op)
	return err
}

// 将源站信息转换为配置
func (this *OriginServerDAO) ComposeOriginConfig(originId int64) (*serverconfigs.OriginServerConfig, error) {
	origin, err := this.FindEnabledOriginServer(originId)
	if err != nil {
		return nil, err
	}
	if origin == nil {
		return nil, errors.New("not found")
	}

	config := &serverconfigs.OriginServerConfig{
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
		Host:         origin.Host,
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

	if origin.RequestHeaderPolicyId > 0 {
		policyConfig, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(int64(origin.RequestHeaderPolicyId))
		if err != nil {
			return nil, err
		}
		if policyConfig != nil {
			config.RequestHeaders = policyConfig
		}
	}

	if origin.ResponseHeaderPolicyId > 0 {
		policyConfig, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(int64(origin.ResponseHeaderPolicyId))
		if err != nil {
			return nil, err
		}
		if policyConfig != nil {
			config.ResponseHeaders = policyConfig
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
		ftp := &serverconfigs.OriginServerFTPConfig{}
		err = json.Unmarshal([]byte(origin.Ftp), ftp)
		if err != nil {
			return nil, err
		}
		config.FTP = ftp
	}

	return config, nil
}
