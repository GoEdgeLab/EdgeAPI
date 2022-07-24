package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPGzipStateEnabled  = 1 // 已启用
	HTTPGzipStateDisabled = 0 // 已禁用
)

type HTTPGzipDAO dbs.DAO

func NewHTTPGzipDAO() *HTTPGzipDAO {
	return dbs.NewDAO(&HTTPGzipDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPGzips",
			Model:  new(HTTPGzip),
			PkName: "id",
		},
	}).(*HTTPGzipDAO)
}

var SharedHTTPGzipDAO *HTTPGzipDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPGzipDAO = NewHTTPGzipDAO()
	})
}

// Init 初始化
func (this *HTTPGzipDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPGzip 启用条目
func (this *HTTPGzipDAO) EnableHTTPGzip(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPGzipStateEnabled).
		Update()
	return err
}

// DisableHTTPGzip 禁用条目
func (this *HTTPGzipDAO) DisableHTTPGzip(tx *dbs.Tx, gzipId int64) error {
	_, err := this.Query(tx).
		Pk(gzipId).
		Set("state", HTTPGzipStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, gzipId)
}

// FindEnabledHTTPGzip 查找启用中的条目
func (this *HTTPGzipDAO) FindEnabledHTTPGzip(tx *dbs.Tx, id int64) (*HTTPGzip, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPGzipStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPGzip), err
}

// ComposeGzipConfig 组合配置
func (this *HTTPGzipDAO) ComposeGzipConfig(tx *dbs.Tx, gzipId int64) (*serverconfigs.HTTPGzipCompressionConfig, error) {
	gzip, err := this.FindEnabledHTTPGzip(tx, gzipId)
	if err != nil {
		return nil, err
	}

	if gzip == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPGzipCompressionConfig{}
	config.Id = int64(gzip.Id)
	config.IsOn = gzip.IsOn
	if IsNotNull(gzip.MinLength) {
		minLengthConfig := &shared.SizeCapacity{}
		err = json.Unmarshal(gzip.MinLength, minLengthConfig)
		if err != nil {
			return nil, err
		}
		config.MinLength = minLengthConfig
	}
	if IsNotNull(gzip.MaxLength) {
		maxLengthConfig := &shared.SizeCapacity{}
		err = json.Unmarshal(gzip.MaxLength, maxLengthConfig)
		if err != nil {
			return nil, err
		}
		config.MaxLength = maxLengthConfig
	}
	config.Level = types.Int8(gzip.Level)

	if IsNotNull(gzip.Conds) {
		condsConfig := &shared.HTTPRequestCondsConfig{}
		err = json.Unmarshal(gzip.Conds, condsConfig)
		if err != nil {
			return nil, err
		}
		config.Conds = condsConfig
	}

	return config, nil
}

// CreateGzip 创建Gzip
func (this *HTTPGzipDAO) CreateGzip(tx *dbs.Tx, level int, minLengthJSON []byte, maxLengthJSON []byte, condsJSON []byte) (int64, error) {
	var op = NewHTTPGzipOperator()
	op.State = HTTPGzipStateEnabled
	op.IsOn = true
	op.Level = level
	if len(minLengthJSON) > 0 {
		op.MinLength = JSONBytes(minLengthJSON)
	}
	if len(maxLengthJSON) > 0 {
		op.MaxLength = JSONBytes(maxLengthJSON)
	}
	if len(condsJSON) > 0 {
		op.Conds = JSONBytes(condsJSON)
	}
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateGzip 修改Gzip
func (this *HTTPGzipDAO) UpdateGzip(tx *dbs.Tx, gzipId int64, level int, minLengthJSON []byte, maxLengthJSON []byte, condsJSON []byte) error {
	if gzipId <= 0 {
		return errors.New("invalid gzipId")
	}
	var op = NewHTTPGzipOperator()
	op.Id = gzipId
	op.Level = level
	if len(minLengthJSON) > 0 {
		op.MinLength = JSONBytes(minLengthJSON)
	}
	if len(maxLengthJSON) > 0 {
		op.MaxLength = JSONBytes(maxLengthJSON)
	}
	if len(condsJSON) > 0 {
		op.Conds = JSONBytes(condsJSON)
	}
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, gzipId)
}

// NotifyUpdate 通知更新
func (this *HTTPGzipDAO) NotifyUpdate(tx *dbs.Tx, gzipId int64) error {
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithGzipId(tx, gzipId)
	if err != nil {
		return err
	}
	if webId > 0 {
		return SharedHTTPWebDAO.NotifyUpdate(tx, webId)
	}
	return nil
}
