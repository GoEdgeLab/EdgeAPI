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

	addr := &serverconfigs.NetworkAddressConfig{}
	if len(origin.Addr) > 0 && origin.Addr != "null" {
		err = json.Unmarshal([]byte(origin.Addr), addr)
		if err != nil {
			return nil, err
		}
	}

	connTimeout := &shared.TimeDuration{}
	if len(origin.ConnTimeout) > 0 && origin.ConnTimeout != "null" {
		err = json.Unmarshal([]byte(origin.ConnTimeout), &connTimeout)
		if err != nil {
			return nil, err
		}
	}

	readTimeout := &shared.TimeDuration{}
	if len(origin.ReadTimeout) > 0 && origin.ReadTimeout != "null" {
		err = json.Unmarshal([]byte(origin.ReadTimeout), &readTimeout)
		if err != nil {
			return nil, err
		}
	}

	idleTimeout := &shared.TimeDuration{}
	if len(origin.IdleTimeout) > 0 && origin.IdleTimeout != "null" {
		err = json.Unmarshal([]byte(origin.IdleTimeout), &idleTimeout)
		if err != nil {
			return nil, err
		}
	}

	requestHeaders := &shared.HTTPHeadersConfig{}
	if len(origin.HttpRequestHeaders) > 0 && origin.HttpRequestHeaders != "null" {
		err = json.Unmarshal([]byte(origin.HttpRequestHeaders), requestHeaders)
		if err != nil {
			return nil, err
		}
	}

	responseHeaders := &shared.HTTPHeadersConfig{}
	if len(origin.HttpResponseHeaders) > 0 && origin.HttpResponseHeaders != "null" {
		err = json.Unmarshal([]byte(origin.HttpResponseHeaders), responseHeaders)
		if err != nil {
			return nil, err
		}
	}

	healthCheck := &serverconfigs.HealthCheckConfig{}
	if len(origin.HealthCheck) > 0 && origin.HealthCheck != "null" {
		err = json.Unmarshal([]byte(origin.HealthCheck), healthCheck)
		if err != nil {
			return nil, err
		}
	}

	cert := &sslconfigs.SSLCertConfig{}
	if len(origin.Cert) > 0 && origin.Cert != "null" {
		err = json.Unmarshal([]byte(origin.Cert), cert)
		if err != nil {
			return nil, err
		}
	}

	ftp := &serverconfigs.OriginServerFTPConfig{}
	if len(origin.Ftp) > 0 && origin.Ftp != "null" {
		err = json.Unmarshal([]byte(origin.Ftp), ftp)
		if err != nil {
			return nil, err
		}
	}

	return &serverconfigs.OriginServerConfig{
		Id:              int64(origin.Id),
		IsOn:            origin.IsOn == 1,
		Version:         int(origin.Version),
		Name:            origin.Name,
		Addr:            addr,
		Description:     origin.Description,
		Code:            origin.Code,
		Weight:          uint(origin.Weight),
		ConnTimeout:     connTimeout,
		ReadTimeout:     readTimeout,
		IdleTimeout:     idleTimeout,
		MaxFails:        int(origin.MaxFails),
		MaxConns:        int(origin.MaxConns),
		MaxIdleConns:    int(origin.MaxIdleConns),
		RequestURI:      origin.HttpRequestURI,
		Host:            origin.Host,
		RequestHeaders:  requestHeaders,
		ResponseHeaders: responseHeaders,
		HealthCheck:     healthCheck,
		Cert:            cert,
		FTP:             ftp,
	}, nil
}
