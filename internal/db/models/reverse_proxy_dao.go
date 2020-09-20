package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
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

var SharedReverseProxyDAO = NewReverseProxyDAO()

// 启用条目
func (this *ReverseProxyDAO) EnableReverseProxy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ReverseProxyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *ReverseProxyDAO) DisableReverseProxy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ReverseProxyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *ReverseProxyDAO) FindEnabledReverseProxy(id int64) (*ReverseProxy, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ReverseProxyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ReverseProxy), err
}

// 根据iD组合配置
func (this *ReverseProxyDAO) ComposeReverseProxyConfig(reverseProxyId int64) (*serverconfigs.ReverseProxyConfig, error) {
	reverseProxy, err := this.FindEnabledReverseProxy(reverseProxyId)
	if err != nil {
		return nil, err
	}
	if reverseProxy == nil {
		return nil, nil
	}

	config := &serverconfigs.ReverseProxyConfig{}
	config.IsOn = reverseProxy.IsOn == 1

	schedulingConfig := &serverconfigs.SchedulingConfig{}
	if len(reverseProxy.Scheduling) > 0 && reverseProxy.Scheduling != "null" {
		err = json.Unmarshal([]byte(reverseProxy.Scheduling), schedulingConfig)
		if err != nil {
			return nil, err
		}
		config.Scheduling = schedulingConfig
	}
	if len(reverseProxy.PrimaryOrigins) > 0 && reverseProxy.PrimaryOrigins != "null" {
		originConfigs := []*serverconfigs.OriginServerConfig{}
		err = json.Unmarshal([]byte(reverseProxy.PrimaryOrigins), &originConfigs)
		if err != nil {
			return nil, err
		}
		for _, originConfig := range originConfigs {
			newOriginConfig, err := SharedOriginServerDAO.ComposeOriginConfig(originConfig.Id)
			if err != nil {
				return nil, err
			}
			if newOriginConfig != nil {
				config.AddPrimaryOrigin(newOriginConfig)
			}
		}
	}

	if len(reverseProxy.BackupOrigins) > 0 && reverseProxy.BackupOrigins != "null" {
		originConfigs := []*serverconfigs.OriginServerConfig{}
		err = json.Unmarshal([]byte(reverseProxy.BackupOrigins), &originConfigs)
		if err != nil {
			return nil, err
		}
		for _, originConfig := range originConfigs {
			newOriginConfig, err := SharedOriginServerDAO.ComposeOriginConfig(int64(originConfig.Id))
			if err != nil {
				return nil, err
			}
			if newOriginConfig != nil {
				config.AddBackupOrigin(newOriginConfig)
			}
		}
	}

	return config, nil
}

// 创建反向代理
func (this *ReverseProxyDAO) CreateReverseProxy(schedulingJSON []byte, primaryOriginsJSON []byte, backupOriginsJSON []byte) (int64, error) {
	op := NewReverseProxyOperator()
	op.State = ReverseProxyStateEnabled
	if len(schedulingJSON) > 0 {
		op.Scheduling = string(schedulingJSON)
	}
	if len(primaryOriginsJSON) > 0 {
		op.PrimaryOrigins = string(primaryOriginsJSON)
	}
	if len(backupOriginsJSON) > 0 {
		op.BackupOrigins = string(backupOriginsJSON)
	}
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// 修改反向代理调度算法
func (this *ReverseProxyDAO) UpdateReverseProxyScheduling(reverseProxyId int64, schedulingJSON []byte) error {
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
	_, err := this.Save(op)

	// TODO 更新所有使用此反向代理的服务

	return err
}

// 修改主要源站
func (this *ReverseProxyDAO) UpdateReverseProxyPrimaryOrigins(reverseProxyId int64, origins []byte) error {
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
	_, err := this.Save(op)

	// TODO 更新所有使用此反向代理的服务

	return err
}

// 修改备用源站
func (this *ReverseProxyDAO) UpdateReverseProxyBackupOrigins(reverseProxyId int64, origins []byte) error {
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
	_, err := this.Save(op)

	// TODO 更新所有使用此反向代理的服务

	return err
}

// 修改是否启用
func (this *ReverseProxyDAO) UpdateReverseProxyIsOn(reverseProxyId int64, isOn bool) error {
	_, err := this.Query().
		Pk(reverseProxyId).
		Set("isOn", isOn).
		Update()
	return err
}