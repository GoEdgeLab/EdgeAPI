package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPFastcgiStateEnabled  = 1 // 已启用
	HTTPFastcgiStateDisabled = 0 // 已禁用
)

type HTTPFastcgiDAO dbs.DAO

func NewHTTPFastcgiDAO() *HTTPFastcgiDAO {
	return dbs.NewDAO(&HTTPFastcgiDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFastcgis",
			Model:  new(HTTPFastcgi),
			PkName: "id",
		},
	}).(*HTTPFastcgiDAO)
}

var SharedHTTPFastcgiDAO *HTTPFastcgiDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPFastcgiDAO = NewHTTPFastcgiDAO()
	})
}

// EnableHTTPFastcgi 启用条目
func (this *HTTPFastcgiDAO) EnableHTTPFastcgi(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPFastcgiStateEnabled).
		Update()
	return err
}

// DisableHTTPFastcgi 禁用条目
func (this *HTTPFastcgiDAO) DisableHTTPFastcgi(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPFastcgiStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPFastcgi 查找启用中的条目
func (this *HTTPFastcgiDAO) FindEnabledHTTPFastcgi(tx *dbs.Tx, id int64) (*HTTPFastcgi, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPFastcgiStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFastcgi), err
}

// ComposeFastcgiConfig 组合配置
func (this *HTTPFastcgiDAO) ComposeFastcgiConfig(tx *dbs.Tx, fastcgiId int64) (*serverconfigs.HTTPFastcgiConfig, error) {
	if fastcgiId <= 0 {
		return nil, nil
	}
	fastcgi, err := this.FindEnabledHTTPFastcgi(tx, fastcgiId)
	if err != nil {
		return nil, err
	}
	if fastcgi == nil {
		return nil, nil
	}
	config := &serverconfigs.HTTPFastcgiConfig{}
	config.Id = int64(fastcgi.Id)
	config.IsOn = fastcgi.IsOn
	config.Address = fastcgi.Address

	if IsNotNull(fastcgi.Params) {
		params := []*serverconfigs.HTTPFastcgiParam{}
		err = json.Unmarshal(fastcgi.Params, &params)
		if err != nil {
			return nil, err
		}
		config.Params = params
	}

	if IsNotNull(fastcgi.ReadTimeout) {
		duration := &shared.TimeDuration{}
		err = json.Unmarshal(fastcgi.ReadTimeout, duration)
		if err != nil {
			return nil, err
		}
		config.ReadTimeout = duration
	}

	if IsNotNull(fastcgi.ConnTimeout) {
		duration := &shared.TimeDuration{}
		err = json.Unmarshal(fastcgi.ConnTimeout, duration)
		if err != nil {
			return nil, err
		}
		config.ConnTimeout = duration
	}

	if fastcgi.PoolSize > 0 {
		config.PoolSize = types.Int(fastcgi.PoolSize)
	}
	config.PathInfoPattern = fastcgi.PathInfoPattern

	return config, nil
}

// CreateFastcgi 创建Fastcgi
func (this *HTTPFastcgiDAO) CreateFastcgi(tx *dbs.Tx, adminId int64, userId int64, isOn bool, address string, paramsJSON []byte, readTimeoutJSON []byte, connTimeoutJSON []byte, poolSize int32, pathInfoPattern string) (int64, error) {
	op := NewHTTPFastcgiOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.IsOn = isOn
	op.Address = address
	if len(paramsJSON) > 0 {
		op.Params = paramsJSON
	}
	if len(readTimeoutJSON) > 0 {
		op.ReadTimeout = readTimeoutJSON
	}
	if len(connTimeoutJSON) > 0 {
		op.ConnTimeout = connTimeoutJSON
	}
	op.PoolSize = poolSize
	op.PathInfoPattern = pathInfoPattern

	op.State = HTTPFastcgiStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateFastcgi 修改Fastcgi
func (this *HTTPFastcgiDAO) UpdateFastcgi(tx *dbs.Tx, fastcgiId int64, isOn bool, address string, paramsJSON []byte, readTimeoutJSON []byte, connTimeoutJSON []byte, poolSize int32, pathInfoPattern string) error {
	if fastcgiId <= 0 {
		return errors.New("invalid 'fastcgiId'")
	}
	op := NewHTTPFastcgiOperator()
	op.Id = fastcgiId
	op.IsOn = isOn
	op.Address = address
	if len(paramsJSON) > 0 {
		op.Params = paramsJSON
	}
	if len(readTimeoutJSON) > 0 {
		op.ReadTimeout = readTimeoutJSON
	}
	if len(connTimeoutJSON) > 0 {
		op.ConnTimeout = connTimeoutJSON
	}
	op.PoolSize = poolSize
	op.PathInfoPattern = pathInfoPattern
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, fastcgiId)
}

// CheckUserFastcgi 检查用户Fastcgi权限
func (this *HTTPFastcgiDAO) CheckUserFastcgi(tx *dbs.Tx, userId int64, fastcgiId int64) error {
	if userId <= 0 || fastcgiId <= 0 {
		return errors.New("permission error")
	}
	exists, err := this.Query(tx).
		Pk(fastcgiId).
		Attr("userId", userId).
		State(HTTPFastcgiStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("permission error")
	}
	return nil
}

// NotifyUpdate 通知更新
func (this *HTTPFastcgiDAO) NotifyUpdate(tx *dbs.Tx, fastcgiId int64) error {
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithFastcgiId(tx, fastcgiId)
	if err != nil {
		return err
	}
	if webId > 0 {
		return SharedHTTPWebDAO.NotifyUpdate(tx, webId)
	}
	return nil
}
