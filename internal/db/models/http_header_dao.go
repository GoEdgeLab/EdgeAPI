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

// Init 初始化
func (this *HTTPHeaderDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPHeader 启用条目
func (this *HTTPHeaderDAO) EnableHTTPHeader(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPHeaderStateEnabled).
		Update()
	return err
}

// DisableHTTPHeader 禁用条目
func (this *HTTPHeaderDAO) DisableHTTPHeader(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPHeaderStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPHeader 查找启用中的条目
func (this *HTTPHeaderDAO) FindEnabledHTTPHeader(tx *dbs.Tx, id int64) (*HTTPHeader, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPHeaderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPHeader), err
}

// FindHTTPHeaderName 根据主键查找名称
func (this *HTTPHeaderDAO) FindHTTPHeaderName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateHeader 创建Header
func (this *HTTPHeaderDAO) CreateHeader(tx *dbs.Tx, name string, value string) (int64, error) {
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

	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateHeader 修改Header
func (this *HTTPHeaderDAO) UpdateHeader(tx *dbs.Tx, headerId int64, name string, value string) error {
	if headerId <= 0 {
		return errors.New("invalid headerId")
	}

	op := NewHTTPHeaderOperator()
	op.Id = headerId
	op.Name = name
	op.Value = value
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, headerId)
}

// ComposeHeaderConfig 组合Header配置
func (this *HTTPHeaderDAO) ComposeHeaderConfig(tx *dbs.Tx, headerId int64) (*shared.HTTPHeaderConfig, error) {
	header, err := this.FindEnabledHTTPHeader(tx, headerId)
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

// NotifyUpdate 通知更新
func (this *HTTPHeaderDAO) NotifyUpdate(tx *dbs.Tx, headerId int64) error {
	policyId, err := SharedHTTPHeaderPolicyDAO.FindHeaderPolicyIdWithHeaderId(tx, headerId)
	if err != nil {
		return err
	}
	if policyId > 0 {
		return SharedHTTPHeaderPolicyDAO.NotifyUpdate(tx, policyId)
	}
	return nil
}
