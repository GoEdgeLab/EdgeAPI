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
	FormalClientBrowserStateEnabled  = 1 // 已启用
	FormalClientBrowserStateDisabled = 0 // 已禁用
)

type FormalClientBrowserDAO dbs.DAO

func NewFormalClientBrowserDAO() *FormalClientBrowserDAO {
	return dbs.NewDAO(&FormalClientBrowserDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeFormalClientBrowsers",
			Model:  new(FormalClientBrowser),
			PkName: "id",
		},
	}).(*FormalClientBrowserDAO)
}

var SharedFormalClientBrowserDAO *FormalClientBrowserDAO

func init() {
	dbs.OnReady(func() {
		SharedFormalClientBrowserDAO = NewFormalClientBrowserDAO()
	})
}

// EnableFormalClientBrowser 启用条目
func (this *FormalClientBrowserDAO) EnableFormalClientBrowser(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", FormalClientBrowserStateEnabled).
		Update()
	return err
}

// DisableFormalClientBrowser 禁用条目
func (this *FormalClientBrowserDAO) DisableFormalClientBrowser(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", FormalClientBrowserStateDisabled).
		Update()
	return err
}

// FindEnabledFormalClientBrowser 查找启用中的条目
func (this *FormalClientBrowserDAO) FindEnabledFormalClientBrowser(tx *dbs.Tx, id int64) (*FormalClientBrowser, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(FormalClientBrowserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*FormalClientBrowser), err
}

// FindFormalClientBrowserName 根据主键查找名称
func (this *FormalClientBrowserDAO) FindFormalClientBrowserName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindBrowserIdWithNameCacheable 根据浏览器名称查找系统ID
func (this *FormalClientBrowserDAO) FindBrowserIdWithNameCacheable(tx *dbs.Tx, browserName string) (int64, error) {
	var cacheKey = "formalClientBrowser:" + browserName
	var cacheItem = ttlcache.SharedCache.Read(cacheKey)
	if cacheItem != nil {
		return types.Int64(cacheItem.Value), nil
	}

	// 先使用 name 查找，因为有索引，所以会快一些
	browserId, err := this.Query(tx).
		Attr("name", browserName).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}

	if browserId == 0 {
		browserId, err = this.Query(tx).
			Where("JSON_CONTAINS(codes, :browserName)").
			Param("browserName", strconv.Quote(browserName)). // 查询的需要是个JSON字符串，所以这里加双引号
			ResultPk().
			FindInt64Col(0)
		if err != nil {
			return 0, err
		}
	}

	// 即使找不到也要放入到缓存中
	ttlcache.SharedCache.Write(cacheKey, browserId, time.Now().Unix()+3600)

	return browserId, nil
}

// CountBrowsers 计算浏览器数量
func (this *FormalClientBrowserDAO) CountBrowsers(tx *dbs.Tx, keyword string) (int64, error) {
	var query = this.Query(tx)
	if len(keyword) > 0 {
		query.Like("LOWER(codes)", dbutils.QuoteLikeKeyword(strings.ToLower(keyword)))
	}
	return query.Count()
}

// ListBrowsers 列出单页浏览器信息
func (this *FormalClientBrowserDAO) ListBrowsers(tx *dbs.Tx, keyword string, offset int64, size int64) (result []*FormalClientBrowser, err error) {
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

// FindBrowserWithDataId 根据dataId查找浏览器信息
func (this *FormalClientBrowserDAO) FindBrowserWithDataId(tx *dbs.Tx, dataId string) (*FormalClientBrowser, error) {
	one, err := this.Query(tx).
		Attr("dataId", dataId).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*FormalClientBrowser), nil
}

// CreateBrowser 创建浏览器信息
func (this *FormalClientBrowserDAO) CreateBrowser(tx *dbs.Tx, name string, codes []string, dataId string) (int64, error) {
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

	var op = NewFormalClientBrowserOperator()
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
	op.State = FormalClientBrowserStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateBrowser 修改浏览器信息
func (this *FormalClientBrowserDAO) UpdateBrowser(tx *dbs.Tx, browserId int64, name string, codes []string, dataId string) error {
	if browserId <= 0 {
		return errors.New("invalid browserId '" + types.String(browserId) + "'")
	}
	if len(dataId) == 0 {
		return errors.New("invalid dataId")
	}

	var op = NewFormalClientBrowserOperator()
	op.Id = browserId
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
