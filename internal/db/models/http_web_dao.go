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
	HTTPWebStateEnabled  = 1 // 已启用
	HTTPWebStateDisabled = 0 // 已禁用
)

type HTTPWebDAO dbs.DAO

func NewHTTPWebDAO() *HTTPWebDAO {
	return dbs.NewDAO(&HTTPWebDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPWebs",
			Model:  new(HTTPWeb),
			PkName: "id",
		},
	}).(*HTTPWebDAO)
}

var SharedHTTPWebDAO = NewHTTPWebDAO()

// 启用条目
func (this *HTTPWebDAO) EnableHTTPWeb(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPWebStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPWebDAO) DisableHTTPWeb(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPWebStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPWebDAO) FindEnabledHTTPWeb(id int64) (*HTTPWeb, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPWebStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPWeb), err
}

// 组合配置
func (this *HTTPWebDAO) ComposeWebConfig(webId int64) (*serverconfigs.HTTPWebConfig, error) {
	web, err := SharedHTTPWebDAO.FindEnabledHTTPWeb(webId)
	if err != nil {
		return nil, err
	}
	if web == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPWebConfig{}
	config.Id = webId
	config.IsOn = web.IsOn == 1
	config.Root = web.Root

	// gzip
	if IsNotNull(web.Gzip) {
		gzipRef := &serverconfigs.HTTPGzipRef{}
		err = json.Unmarshal([]byte(web.Gzip), gzipRef)
		if err != nil {
			return nil, err
		}
		config.GzipRef = gzipRef

		gzipConfig, err := SharedHTTPGzipDAO.ComposeGzipConfig(gzipRef.GzipId)
		if err != nil {
			return nil, err
		}
		config.Gzip = gzipConfig
	}

	// charset
	config.Charset = web.Charset

	// headers
	if web.RequestHeaderPolicyId > 0 {
		headerPolicy, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(int64(web.RequestHeaderPolicyId))
		if err != nil {
			return nil, err
		}
		if headerPolicy != nil {
			config.RequestHeaders = headerPolicy
		}
	}

	if web.ResponseHeaderPolicyId > 0 {
		headerPolicy, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(int64(web.ResponseHeaderPolicyId))
		if err != nil {
			return nil, err
		}
		if headerPolicy != nil {
			config.ResponseHeaders = headerPolicy
		}
	}

	// shutdown
	if IsNotNull(web.Shutdown) {
		shutdownConfig := &serverconfigs.HTTPShutdownConfig{}
		err = json.Unmarshal([]byte(web.Shutdown), shutdownConfig)
		if err != nil {
			return nil, err
		}
		config.Shutdown = shutdownConfig
	}

	// pages
	if IsNotNull(web.Pages) {
		pages := []*serverconfigs.HTTPPageConfig{}
		err = json.Unmarshal([]byte(web.Pages), &pages)
		if err != nil {
			return nil, err
		}
		for index, page := range pages {
			pageConfig, err := SharedHTTPPageDAO.ComposePageConfig(page.Id)
			if err != nil {
				return nil, err
			}
			pages[index] = pageConfig
		}
		if len(pages) > 0 {
			config.Pages = pages
		}
	}

	// 访问日志
	if IsNotNull(web.AccessLog) {
		accessLogConfig := &serverconfigs.HTTPAccessLogRef{}
		err = json.Unmarshal([]byte(web.AccessLog), accessLogConfig)
		if err != nil {
			return nil, err
		}
		config.AccessLogRef = accessLogConfig
	}

	// 统计配置
	if IsNotNull(web.Stat) {
		statRef := &serverconfigs.HTTPStatRef{}
		err = json.Unmarshal([]byte(web.Stat), statRef)
		if err != nil {
			return nil, err
		}
		config.StatRef = statRef
	}

	// 缓存配置
	if IsNotNull(web.Cache) {
		cacheRef := &serverconfigs.HTTPCacheRef{}
		err = json.Unmarshal([]byte(web.Cache), cacheRef)
		if err != nil {
			return nil, err
		}
		config.CacheRef = cacheRef
	}

	// 防火墙配置
	if IsNotNull(web.Firewall) {
		firewallRef := &serverconfigs.HTTPFirewallRef{}
		err = json.Unmarshal([]byte(web.Firewall), firewallRef)
		if err != nil {
			return nil, err
		}
		config.FirewallRef = firewallRef
	}

	// 路径规则
	if IsNotNull(web.Locations) {
		refs := []*serverconfigs.HTTPLocationRef{}
		err = json.Unmarshal([]byte(web.Locations), &refs)
		if err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			config.LocationRefs = refs

			locations, err := SharedHTTPLocationDAO.ConvertLocationRefs(refs)
			if err != nil {
				return nil, err
			}
			config.Locations = locations
		}
	}

	// TODO 更多配置

	return config, nil
}

// 创建Web配置
func (this *HTTPWebDAO) CreateWeb(root string) (int64, error) {
	op := NewHTTPWebOperator()
	op.State = HTTPWebStateEnabled
	op.Root = root
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改Web配置
func (this *HTTPWebDAO) UpdateWeb(webId int64, root string) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Root = root
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 修改Gzip配置
func (this *HTTPWebDAO) UpdateWebGzip(webId int64, gzipJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Gzip = gzipJSON
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 修改字符编码
func (this *HTTPWebDAO) UpdateWebCharset(webId int64, charset string) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Charset = charset
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改请求Header策略
func (this *HTTPWebDAO) UpdateWebRequestHeaderPolicy(webId int64, headerPolicyId int64) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.RequestHeaderPolicyId = headerPolicyId
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改响应Header策略
func (this *HTTPWebDAO) UpdateWebResponseHeaderPolicy(webId int64, headerPolicyId int64) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.ResponseHeaderPolicyId = headerPolicyId
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改特殊页面配置
func (this *HTTPWebDAO) UpdateWebPages(webId int64, pagesJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Pages = JSONBytes(pagesJSON)
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改Shutdown配置
func (this *HTTPWebDAO) UpdateWebShutdown(webId int64, shutdownJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Shutdown = JSONBytes(shutdownJSON)
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改访问日志策略
func (this *HTTPWebDAO) UpdateWebAccessLogConfig(webId int64, accessLogJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.AccessLog = JSONBytes(accessLogJSON)
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改统计配置
func (this *HTTPWebDAO) UpdateWebStat(webId int64, statJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Stat = JSONBytes(statJSON)
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改缓存配置
func (this *HTTPWebDAO) UpdateWebCache(webId int64, cacheJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Cache = JSONBytes(cacheJSON)
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改防火墙配置
func (this *HTTPWebDAO) UpdateWebFirewall(webId int64, firewallJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Firewall = JSONBytes(firewallJSON)
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 更改路径规则配置
func (this *HTTPWebDAO) UpdateWebLocations(webId int64, locationsJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Locations = JSONBytes(locationsJSON)
	_, err := this.Save(op)
	if err != nil {
		return err
	}

	return this.NotifyUpdating(webId)
}

// 通知更新
func (this *HTTPWebDAO) NotifyUpdating(webId int64) error {
	err := SharedServerDAO.UpdateServerIsUpdatingWithWebId(webId)
	if err != nil {
		return err
	}

	// TODO 更新所有使用此Web配置的Location所在服务

	return nil
}
