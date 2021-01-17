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

// 初始化
func (this *HTTPPageDAO) Init() {
	_ = this.DAOObject.Init()
}

// 启用条目
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

// 禁用条目
func (this *HTTPPageDAO) DisableHTTPPage(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPPageStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
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

// 创建Page
func (this *HTTPPageDAO) CreatePage(tx *dbs.Tx, statusList []string, url string, newStatus int) (pageId int64, err error) {
	op := NewHTTPPageOperator()
	op.IsOn = true
	op.State = HTTPPageStateEnabled

	if len(statusList) > 0 {
		statusListJSON, err := json.Marshal(statusList)
		if err != nil {
			return 0, err
		}
		op.StatusList = string(statusListJSON)
	}
	op.Url = url
	op.NewStatus = newStatus
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// 修改Page
func (this *HTTPPageDAO) UpdatePage(tx *dbs.Tx, pageId int64, statusList []string, url string, newStatus int) error {
	if pageId <= 0 {
		return errors.New("invalid pageId")
	}

	op := NewHTTPPageOperator()
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

	op.Url = url
	op.NewStatus = newStatus
	err = this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, pageId)
}

// 组合配置
func (this *HTTPPageDAO) ComposePageConfig(tx *dbs.Tx, pageId int64) (*serverconfigs.HTTPPageConfig, error) {
	page, err := this.FindEnabledHTTPPage(tx, pageId)
	if err != nil {
		return nil, err
	}

	if page == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPPageConfig{}
	config.Id = int64(page.Id)
	config.IsOn = page.IsOn == 1
	config.NewStatus = int(page.NewStatus)
	config.URL = page.Url

	if len(page.StatusList) > 0 {
		statusList := []string{}
		err = json.Unmarshal([]byte(page.StatusList), &statusList)
		if err != nil {
			return nil, err
		}
		if len(statusList) > 0 {
			config.Status = statusList
		}
	}

	return config, nil
}

// 通知更新
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
