package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPHeaderStateEnabled  = 1 // 已启用
	HTTPHeaderStateDisabled = 0 // 已禁用
)

type HTTPHeaderDAO dbs.DAO

func NewHTTPHeaderDAO() *HTTPHeaderDAO {
	return dbs.NewDAO(&HTTPHeaderDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPHeaders",
			Model:  new(HTTPHeader),
			PkName: "id",
		},
	}).(*HTTPHeaderDAO)
}

var SharedHTTPHeaderDAO *HTTPHeaderDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPHeaderDAO = NewHTTPHeaderDAO()
	})
}

// 初始化
func (this *HTTPHeaderDAO) Init() {
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
func (this *HTTPHeaderDAO) EnableHTTPHeader(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPHeaderStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPHeaderDAO) DisableHTTPHeader(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPHeaderStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPHeaderDAO) FindEnabledHTTPHeader(id int64) (*HTTPHeader, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPHeaderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPHeader), err
}

// 根据主键查找名称
func (this *HTTPHeaderDAO) FindHTTPHeaderName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建Header
func (this *HTTPHeaderDAO) CreateHeader(name string, value string) (int64, error) {
	op := NewHTTPHeaderOperator()
	op.State = HTTPHeaderStateEnabled
	op.IsOn = true
	op.Name = name
	op.Value = value

	statusConfig := &shared.HTTPStatusConfig{
		Always: true,
	}
	statusJSON, err := json.Marshal(statusConfig)
	if err != nil {
		return 0, err
	}
	op.Status = statusJSON

	err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改Header
func (this *HTTPHeaderDAO) UpdateHeader(headerId int64, name string, value string) error {
	if headerId <= 0 {
		return errors.New("invalid headerId")
	}

	op := NewHTTPHeaderOperator()
	op.Id = headerId
	op.Name = name
	op.Value = value
	err := this.Save(op)

	// TODO 更新相关配置

	return err
}

// 组合Header配置
func (this *HTTPHeaderDAO) ComposeHeaderConfig(headerId int64) (*shared.HTTPHeaderConfig, error) {
	header, err := this.FindEnabledHTTPHeader(headerId)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, nil
	}

	config := &shared.HTTPHeaderConfig{}
	config.Id = int64(header.Id)
	config.IsOn = header.IsOn == 1
	config.Name = header.Name
	config.Value = header.Value

	if len(header.Status) > 0 {
		status := &shared.HTTPStatusConfig{}
		err = json.Unmarshal([]byte(header.Status), status)
		if err != nil {
			return nil, err
		}
		config.Status = status
	}

	return config, nil
}
