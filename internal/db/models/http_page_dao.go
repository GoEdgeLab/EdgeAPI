package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPPageStateEnabled  = 1 // 已启用
	HTTPPageStateDisabled = 0 // 已禁用
)

type HTTPPageDAO dbs.DAO

func NewHTTPPageDAO() *HTTPPageDAO {
	return dbs.NewDAO(&HTTPPageDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPPages",
			Model:  new(HTTPPage),
			PkName: "id",
		},
	}).(*HTTPPageDAO)
}

var SharedHTTPPageDAO *HTTPPageDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPPageDAO = NewHTTPPageDAO()
	})
}

// Init 初始化
func (this *HTTPPageDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPPage 启用条目
func (this *HTTPPageDAO) EnableHTTPPage(tx *dbs.Tx, pageId int64) error {
	_, err := this.Query(tx).
		Pk(pageId).
		Set("state", HTTPPageStateEnabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, pageId)
}

// DisableHTTPPage 禁用条目
func (this *HTTPPageDAO) DisableHTTPPage(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPPageStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPPage 查找启用中的条目
func (this *HTTPPageDAO) FindEnabledHTTPPage(tx *dbs.Tx, id int64) (*HTTPPage, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPPageStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPPage), err
}

// CreatePage 创建Page
func (this *HTTPPageDAO) CreatePage(tx *dbs.Tx, userId int64, statusList []string, bodyType serverconfigs.HTTPPageBodyType, url string, body string, newStatus int, exceptURLPatterns []*shared.URLPattern, onlyURLPatterns []*shared.URLPattern) (pageId int64, err error) {
	var op = NewHTTPPageOperator()
	op.UserId = userId
	op.IsOn = true
	op.State = HTTPPageStateEnabled

	if len(statusList) > 0 {
		statusListJSON, err := json.Marshal(statusList)
		if err != nil {
			return 0, err
		}
		op.StatusList = string(statusListJSON)
	}
	op.BodyType = bodyType
	op.Url = url
	op.Body = body
	op.NewStatus = newStatus

	{
		if exceptURLPatterns == nil {
			exceptURLPatterns = []*shared.URLPattern{}
		}
		exceptURLPatternsJSON, err := json.Marshal(exceptURLPatterns)
		if err != nil {
			return 0, err
		}
		op.ExceptURLPatterns = exceptURLPatternsJSON
	}

	{
		if onlyURLPatterns == nil {
			onlyURLPatterns = []*shared.URLPattern{}
		}
		onlyURLPatternsJSON, err := json.Marshal(onlyURLPatterns)
		if err != nil {
			return 0, err
		}
		op.OnlyURLPatterns = onlyURLPatternsJSON
	}

	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// UpdatePage 修改Page
func (this *HTTPPageDAO) UpdatePage(tx *dbs.Tx, pageId int64, statusList []string, bodyType serverconfigs.HTTPPageBodyType, url string, body string, newStatus int, exceptURLPatterns []*shared.URLPattern, onlyURLPatterns []*shared.URLPattern) error {
	if pageId <= 0 {
		return errors.New("invalid pageId")
	}

	var op = NewHTTPPageOperator()
	op.Id = pageId
	op.IsOn = true
	op.State = HTTPPageStateEnabled

	if statusList == nil {
		statusList = []string{}
	}
	statusListJSON, err := json.Marshal(statusList)
	if err != nil {
		return err
	}
	op.StatusList = string(statusListJSON)

	op.BodyType = bodyType
	op.Url = url
	op.Body = body
	op.NewStatus = newStatus

	{
		if exceptURLPatterns == nil {
			exceptURLPatterns = []*shared.URLPattern{}
		}
		exceptURLPatternsJSON, err := json.Marshal(exceptURLPatterns)
		if err != nil {
			return err
		}
		op.ExceptURLPatterns = exceptURLPatternsJSON
	}

	{
		if onlyURLPatterns == nil {
			onlyURLPatterns = []*shared.URLPattern{}
		}
		onlyURLPatternsJSON, err := json.Marshal(onlyURLPatterns)
		if err != nil {
			return err
		}
		op.OnlyURLPatterns = onlyURLPatternsJSON
	}

	err = this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, pageId)
}

// ClonePage 克隆页面
func (this *HTTPPageDAO) ClonePage(tx *dbs.Tx, fromPageId int64) (newPageId int64, err error) {
	if fromPageId <= 0 {
		return
	}
	pageOne, err := this.Query(tx).
		Pk(fromPageId).
		Find()
	if err != nil || pageOne == nil {
		return 0, err
	}
	var page = pageOne.(*HTTPPage)

	var op = NewHTTPPageOperator()
	op.IsOn = page.IsOn
	if len(page.StatusList) > 0 {
		op.StatusList = page.StatusList
	}
	op.Url = page.Url
	op.NewStatus = page.NewStatus
	op.Body = page.Body
	op.BodyType = page.BodyType
	op.State = page.State

	if len(page.ExceptURLPatterns) > 0 {
		op.ExceptURLPatterns = page.ExceptURLPatterns
	}
	if len(page.OnlyURLPatterns) > 0 {
		op.OnlyURLPatterns = page.OnlyURLPatterns
	}

	return this.SaveInt64(tx, op)
}

// ComposePageConfig 组合配置
func (this *HTTPPageDAO) ComposePageConfig(tx *dbs.Tx, pageId int64, cacheMap *utils.CacheMap) (*serverconfigs.HTTPPageConfig, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":config:" + types.String(pageId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*serverconfigs.HTTPPageConfig), nil
	}

	page, err := this.FindEnabledHTTPPage(tx, pageId)
	if err != nil {
		return nil, err
	}

	if page == nil {
		return nil, nil
	}

	var config = &serverconfigs.HTTPPageConfig{}
	config.Id = int64(page.Id)
	config.IsOn = page.IsOn
	config.NewStatus = int(page.NewStatus)
	config.URL = page.Url
	config.Body = page.Body
	config.BodyType = page.BodyType

	if len(page.BodyType) == 0 {
		page.BodyType = serverconfigs.HTTPPageBodyTypeURL
	}

	if len(page.StatusList) > 0 {
		statusList := []string{}
		err = json.Unmarshal(page.StatusList, &statusList)
		if err != nil {
			return nil, err
		}
		if len(statusList) > 0 {
			config.Status = statusList
		}
	}

	if len(page.ExceptURLPatterns) > 0 {
		var exceptURLPatterns = []*shared.URLPattern{}
		err = json.Unmarshal(page.ExceptURLPatterns, &exceptURLPatterns)
		if err != nil {
			return nil, err
		}
		if len(exceptURLPatterns) > 0 {
			config.ExceptURLPatterns = exceptURLPatterns
		}
	}

	if len(page.OnlyURLPatterns) > 0 {
		var onlyURLPatterns = []*shared.URLPattern{}
		err = json.Unmarshal(page.OnlyURLPatterns, &onlyURLPatterns)
		if err != nil {
			return nil, err
		}
		if len(onlyURLPatterns) > 0 {
			config.OnlyURLPatterns = onlyURLPatterns
		}
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// CheckUserPage 检查用户页面
func (this *HTTPPageDAO) CheckUserPage(tx *dbs.Tx, userId int64, pageId int64) error {
	if userId <= 0 || pageId <= 0 {
		return ErrNotFound
	}

	b, err := this.Query(tx).
		Pk(pageId).
		Attr("userId", userId).
		State(HTTPPageStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !b {
		return ErrNotFound
	}
	return nil
}

// NotifyUpdate 通知更新
func (this *HTTPPageDAO) NotifyUpdate(tx *dbs.Tx, pageId int64) error {
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithPageId(tx, pageId)
	if err != nil {
		return err
	}
	if webId > 0 {
		return SharedHTTPWebDAO.NotifyUpdate(tx, webId)
	}
	return nil
}
