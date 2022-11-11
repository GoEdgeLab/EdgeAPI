package models

import (
	"encoding/json"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/ttlcache"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strconv"
	"strings"
	"time"
)

const (
	FormalClientSystemStateEnabled  = 1 // 已启用
	FormalClientSystemStateDisabled = 0 // 已禁用
)

type FormalClientSystemDAO dbs.DAO

func NewFormalClientSystemDAO() *FormalClientSystemDAO {
	return dbs.NewDAO(&FormalClientSystemDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeFormalClientSystems",
			Model:  new(FormalClientSystem),
			PkName: "id",
		},
	}).(*FormalClientSystemDAO)
}

var SharedFormalClientSystemDAO *FormalClientSystemDAO

func init() {
	dbs.OnReady(func() {
		SharedFormalClientSystemDAO = NewFormalClientSystemDAO()
	})
}

// EnableFormalClientSystem 启用条目
func (this *FormalClientSystemDAO) EnableFormalClientSystem(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", FormalClientSystemStateEnabled).
		Update()
	return err
}

// DisableFormalClientSystem 禁用条目
func (this *FormalClientSystemDAO) DisableFormalClientSystem(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", FormalClientSystemStateDisabled).
		Update()
	return err
}

// FindEnabledFormalClientSystem 查找启用中的条目
func (this *FormalClientSystemDAO) FindEnabledFormalClientSystem(tx *dbs.Tx, id int64) (*FormalClientSystem, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(FormalClientSystemStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*FormalClientSystem), err
}

// FindFormalClientSystemName 根据主键查找名称
func (this *FormalClientSystemDAO) FindFormalClientSystemName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindSystemIdWithNameCacheable 根据操作系统名称查找系统ID
func (this *FormalClientSystemDAO) FindSystemIdWithNameCacheable(tx *dbs.Tx, systemName string) (int64, error) {
	var cacheKey = "formalClientSystem:" + systemName
	var cacheItem = ttlcache.SharedCache.Read(cacheKey)
	if cacheItem != nil {
		return types.Int64(cacheItem.Value), nil
	}

	// 先使用 name 查找，因为有索引，所以会快一些
	systemId, err := this.Query(tx).
		Attr("name", systemName).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}

	if systemId == 0 {
		systemId, err = this.Query(tx).
			Where("JSON_CONTAINS(codes, :systemName)").
			Param("systemName", strconv.Quote(systemName)). // 查询的需要是个JSON字符串，所以这里加双引号
			ResultPk().
			FindInt64Col(0)
		if err != nil {
			return 0, err
		}
	}

	// 即使找不到也要放入到缓存中
	ttlcache.SharedCache.Write(cacheKey, systemId, time.Now().Unix()+3600)

	return systemId, nil
}

// CreateSystem 创建操作系统信息
func (this *FormalClientSystemDAO) CreateSystem(tx *dbs.Tx, name string, codes []string, dataId string) (int64, error) {
	if len(dataId) == 0 {
		return 0, errors.New("invalid dataId")
	}

	// 检查 dataId 是否已经存在
	exists, err := this.Query(tx).
		Attr("dataId", dataId).
		Exist()
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("dataId '" + dataId + "' already exists")
	}

	var op = NewFormalClientSystemOperator()
	op.Name = name
	if len(codes) == 0 {
		op.Codes = "[]"
	} else {
		codesJSON, err := json.Marshal(codes)
		if err != nil {
			return 0, err
		}
		op.Codes = codesJSON
	}
	op.DataId = dataId
	op.State = FormalClientSystemStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateSystem 修改操作系统信息
func (this *FormalClientSystemDAO) UpdateSystem(tx *dbs.Tx, systemId int64, name string, codes []string, dataId string) error {
	if systemId <= 0 {
		return errors.New("invalid systemId '" + types.String(systemId) + "'")
	}
	if len(dataId) == 0 {
		return errors.New("invalid dataId")
	}

	var op = NewFormalClientSystemOperator()
	op.Id = systemId
	op.Name = name
	if len(codes) == 0 {
		op.Codes = "[]"
	} else {
		codesJSON, err := json.Marshal(codes)
		if err != nil {
			return err
		}
		op.Codes = codesJSON
	}
	op.DataId = dataId
	return this.Save(tx, op)
}

// CountSystems 计算操作系统数量
func (this *FormalClientSystemDAO) CountSystems(tx *dbs.Tx, keyword string) (int64, error) {
	var query = this.Query(tx)
	if len(keyword) > 0 {
		query.Like("LOWER(codes)", dbutils.QuoteLikeKeyword(strings.ToLower(keyword)))
	}
	return query.Count()
}

// ListSystems 列出单页操作系统信息
func (this *FormalClientSystemDAO) ListSystems(tx *dbs.Tx, keyword string, offset int64, size int64) (result []*FormalClientSystem, err error) {
	var query = this.Query(tx)
	if len(keyword) > 0 {
		query.Like("LOWER(codes)", dbutils.QuoteLikeKeyword(strings.ToLower(keyword)))
	}
	_, err = query.
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindSystemWithDataId 根据dataId查找操作系统信息
func (this *FormalClientSystemDAO) FindSystemWithDataId(tx *dbs.Tx, dataId string) (*FormalClientSystem, error) {
	one, err := this.Query(tx).
		Attr("dataId", dataId).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*FormalClientSystem), nil
}
