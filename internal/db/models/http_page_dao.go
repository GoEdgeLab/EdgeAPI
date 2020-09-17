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

var SharedHTTPPageDAO = NewHTTPPageDAO()

// 启用条目
func (this *HTTPPageDAO) EnableHTTPPage(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPPageStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPPageDAO) DisableHTTPPage(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPPageStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPPageDAO) FindEnabledHTTPPage(id int64) (*HTTPPage, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPPageStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPPage), err
}

// 创建Page
func (this *HTTPPageDAO) CreatePage(statusList []string, url string, newStatus int) (pageId int64, err error) {
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
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// 修改Page
func (this *HTTPPageDAO) UpdatePage(pageId int64, statusList []string, url string, newStatus int) error {
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
	_, err = this.Save(op)

	// TODO 修改相关引用的对象

	return err
}

// 组合配置
func (this *HTTPPageDAO) ComposePageConfig(pageId int64) (*serverconfigs.HTTPPageConfig, error) {
	page, err := this.FindEnabledHTTPPage(pageId)
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
