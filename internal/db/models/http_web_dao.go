package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"strconv"
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

var SharedHTTPWebDAO *HTTPWebDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPWebDAO = NewHTTPWebDAO()
	})
}

func (this *HTTPWebDAO) Init() {
	this.DAOObject.Init()
	this.DAOObject.OnUpdate(func() error {
		return SharedSysEventDAO.CreateEvent(nil, NewServerChangeEvent())
	})
	this.DAOObject.OnInsert(func() error {
		return SharedSysEventDAO.CreateEvent(nil, NewServerChangeEvent())
	})
	this.DAOObject.OnDelete(func() error {
		return SharedSysEventDAO.CreateEvent(nil, NewServerChangeEvent())
	})
}

// 启用条目
func (this *HTTPWebDAO) EnableHTTPWeb(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPWebStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPWebDAO) DisableHTTPWeb(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPWebStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPWebDAO) FindEnabledHTTPWeb(tx *dbs.Tx, id int64) (*HTTPWeb, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPWebStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPWeb), err
}

// 组合配置
func (this *HTTPWebDAO) ComposeWebConfig(tx *dbs.Tx, webId int64) (*serverconfigs.HTTPWebConfig, error) {
	web, err := SharedHTTPWebDAO.FindEnabledHTTPWeb(tx, webId)
	if err != nil {
		return nil, err
	}
	if web == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPWebConfig{}
	config.Id = webId
	config.IsOn = web.IsOn == 1

	// root
	if IsNotNull(web.Root) {
		rootConfig := &serverconfigs.HTTPRootConfig{}
		err = json.Unmarshal([]byte(web.Root), rootConfig)
		if err != nil {
			return nil, err
		}
		config.Root = rootConfig
	}

	// gzip
	if IsNotNull(web.Gzip) {
		gzipRef := &serverconfigs.HTTPGzipRef{}
		err = json.Unmarshal([]byte(web.Gzip), gzipRef)
		if err != nil {
			return nil, err
		}
		config.GzipRef = gzipRef

		gzipConfig, err := SharedHTTPGzipDAO.ComposeGzipConfig(tx, gzipRef.GzipId)
		if err != nil {
			return nil, err
		}
		config.Gzip = gzipConfig
	}

	// charset
	if IsNotNull(web.Charset) {
		charsetConfig := &serverconfigs.HTTPCharsetConfig{}
		err = json.Unmarshal([]byte(web.Charset), charsetConfig)
		if err != nil {
			return nil, err
		}
		config.Charset = charsetConfig
	}

	// headers
	if IsNotNull(web.RequestHeader) {
		ref := &shared.HTTPHeaderPolicyRef{}
		err = json.Unmarshal([]byte(web.RequestHeader), ref)
		if err != nil {
			return nil, err
		}
		config.RequestHeaderPolicyRef = ref

		if ref.HeaderPolicyId > 0 {
			headerPolicy, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(tx, ref.HeaderPolicyId)
			if err != nil {
				return nil, err
			}
			if headerPolicy != nil {
				config.RequestHeaderPolicy = headerPolicy
			}
		}
	}

	if IsNotNull(web.ResponseHeader) {
		ref := &shared.HTTPHeaderPolicyRef{}
		err = json.Unmarshal([]byte(web.ResponseHeader), ref)
		if err != nil {
			return nil, err
		}
		config.ResponseHeaderPolicyRef = ref

		if ref.HeaderPolicyId > 0 {
			headerPolicy, err := SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(tx, ref.HeaderPolicyId)
			if err != nil {
				return nil, err
			}
			if headerPolicy != nil {
				config.ResponseHeaderPolicy = headerPolicy
			}
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
			pageConfig, err := SharedHTTPPageDAO.ComposePageConfig(tx, page.Id)
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
		cacheConfig := &serverconfigs.HTTPCacheConfig{}
		err = json.Unmarshal([]byte(web.Cache), &cacheConfig)
		if err != nil {
			return nil, err
		}
		config.Cache = cacheConfig

		// 暂不支持自定义缓存策略设置，因为同一个集群下的服务需要集中管理
	}

	// 防火墙配置
	if IsNotNull(web.Firewall) {
		firewallRef := &firewallconfigs.HTTPFirewallRef{}
		err = json.Unmarshal([]byte(web.Firewall), firewallRef)
		if err != nil {
			return nil, err
		}
		config.FirewallRef = firewallRef

		// 自定义防火墙设置
		if firewallRef.FirewallPolicyId > 0 {
			firewallPolicy, err := SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, firewallRef.FirewallPolicyId)
			if err != nil {
				return nil, err
			}
			if firewallPolicy == nil {
				config.FirewallRef = nil
			} else {
				config.FirewallPolicy = firewallPolicy
			}
		}
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

			locations, err := SharedHTTPLocationDAO.ConvertLocationRefs(tx, refs)
			if err != nil {
				return nil, err
			}
			config.Locations = locations
		}
	}

	// 跳转
	if IsNotNull(web.RedirectToHttps) {
		redirectToHTTPSConfig := &serverconfigs.HTTPRedirectToHTTPSConfig{}
		err = json.Unmarshal([]byte(web.RedirectToHttps), redirectToHTTPSConfig)
		if err != nil {
			return nil, err
		}
		config.RedirectToHttps = redirectToHTTPSConfig
	}

	// Websocket
	if IsNotNull(web.Websocket) {
		ref := &serverconfigs.HTTPWebsocketRef{}
		err = json.Unmarshal([]byte(web.Websocket), ref)
		if err != nil {
			return nil, err
		}
		config.WebsocketRef = ref
		if ref.WebsocketId > 0 {
			websocketConfig, err := SharedHTTPWebsocketDAO.ComposeWebsocketConfig(tx, ref.WebsocketId)
			if err != nil {
				return nil, err
			}
			if websocketConfig != nil {
				config.Websocket = websocketConfig
			}
		}
	}

	// 重写规则
	if IsNotNull(web.RewriteRules) {
		refs := []*serverconfigs.HTTPRewriteRef{}
		err = json.Unmarshal([]byte(web.RewriteRules), &refs)
		if err != nil {
			return nil, err
		}
		for _, ref := range refs {
			rewriteRule, err := SharedHTTPRewriteRuleDAO.ComposeRewriteRule(tx, ref.RewriteRuleId)
			if err != nil {
				return nil, err
			}
			if rewriteRule != nil {
				config.RewriteRefs = append(config.RewriteRefs, ref)
				config.RewriteRules = append(config.RewriteRules, rewriteRule)
			}
		}
	}

	// 主机跳转
	if IsNotNull(web.HostRedirects) {
		redirects := []*serverconfigs.HTTPHostRedirectConfig{}
		err = json.Unmarshal([]byte(web.HostRedirects), &redirects)
		if err != nil {
			return nil, err
		}
		config.HostRedirects = redirects
	}

	return config, nil
}

// 创建Web配置
func (this *HTTPWebDAO) CreateWeb(tx *dbs.Tx, adminId int64, userId int64, rootJSON []byte) (int64, error) {
	op := NewHTTPWebOperator()
	op.State = HTTPWebStateEnabled
	op.AdminId = adminId
	op.UserId = userId
	if len(rootJSON) > 0 {
		op.Root = JSONBytes(rootJSON)
	}
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改Web配置
func (this *HTTPWebDAO) UpdateWeb(tx *dbs.Tx, webId int64, rootJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Root = JSONBytes(rootJSON)
	err := this.Save(tx, op)
	return err
}

// 修改Gzip配置
func (this *HTTPWebDAO) UpdateWebGzip(tx *dbs.Tx, webId int64, gzipJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Gzip = JSONBytes(gzipJSON)
	err := this.Save(tx, op)
	return err
}

// 修改字符编码
func (this *HTTPWebDAO) UpdateWebCharset(tx *dbs.Tx, webId int64, charsetJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Charset = JSONBytes(charsetJSON)
	err := this.Save(tx, op)
	return err
}

// 更改请求Header策略
func (this *HTTPWebDAO) UpdateWebRequestHeaderPolicy(tx *dbs.Tx, webId int64, headerPolicyJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.RequestHeader = JSONBytes(headerPolicyJSON)
	err := this.Save(tx, op)
	return err
}

// 更改响应Header策略
func (this *HTTPWebDAO) UpdateWebResponseHeaderPolicy(tx *dbs.Tx, webId int64, headerPolicyJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.ResponseHeader = JSONBytes(headerPolicyJSON)
	err := this.Save(tx, op)
	return err
}

// 更改特殊页面配置
func (this *HTTPWebDAO) UpdateWebPages(tx *dbs.Tx, webId int64, pagesJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Pages = JSONBytes(pagesJSON)
	err := this.Save(tx, op)
	return err
}

// 更改Shutdown配置
func (this *HTTPWebDAO) UpdateWebShutdown(tx *dbs.Tx, webId int64, shutdownJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Shutdown = JSONBytes(shutdownJSON)
	err := this.Save(tx, op)
	return err
}

// 更改访问日志策略
func (this *HTTPWebDAO) UpdateWebAccessLogConfig(tx *dbs.Tx, webId int64, accessLogJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.AccessLog = JSONBytes(accessLogJSON)
	err := this.Save(tx, op)
	return err
}

// 更改统计配置
func (this *HTTPWebDAO) UpdateWebStat(tx *dbs.Tx, webId int64, statJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Stat = JSONBytes(statJSON)
	err := this.Save(tx, op)
	return err
}

// 更改缓存配置
func (this *HTTPWebDAO) UpdateWebCache(tx *dbs.Tx, webId int64, cacheJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Cache = JSONBytes(cacheJSON)
	err := this.Save(tx, op)
	return err
}

// 更改防火墙配置
func (this *HTTPWebDAO) UpdateWebFirewall(tx *dbs.Tx, webId int64, firewallJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Firewall = JSONBytes(firewallJSON)
	err := this.Save(tx, op)
	return err
}

// 更改路径规则配置
func (this *HTTPWebDAO) UpdateWebLocations(tx *dbs.Tx, webId int64, locationsJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Locations = JSONBytes(locationsJSON)
	err := this.Save(tx, op)
	return err
}

// 更改跳转到HTTPS设置
func (this *HTTPWebDAO) UpdateWebRedirectToHTTPS(tx *dbs.Tx, webId int64, redirectToHTTPSJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.RedirectToHttps = JSONBytes(redirectToHTTPSJSON)
	err := this.Save(tx, op)
	return err
}

// 修改Websocket设置
func (this *HTTPWebDAO) UpdateWebsocket(tx *dbs.Tx, webId int64, websocketJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Websocket = JSONBytes(websocketJSON)
	err := this.Save(tx, op)
	return err
}

// 修改重写规则设置
func (this *HTTPWebDAO) UpdateWebRewriteRules(tx *dbs.Tx, webId int64, rewriteRulesJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.RewriteRules = JSONBytes(rewriteRulesJSON)
	err := this.Save(tx, op)
	return err
}

// 根据缓存策略ID查找所有的WebId
func (this *HTTPWebDAO) FindAllWebIdsWithCachePolicyId(tx *dbs.Tx, cachePolicyId int64) ([]int64, error) {
	ones, err := this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where(`JSON_CONTAINS(cache, '{"cachePolicyId": ` + strconv.FormatInt(cachePolicyId, 10) + ` }', '$.cacheRefs')`).
		Reuse(false). // 由于我们在JSON_CONTAINS()直接使用了变量，所以不能重用
		FindAll()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		webId := int64(one.(*HTTPWeb).Id)

		// 判断是否为Location
		for {
			locationId, err := SharedHTTPLocationDAO.FindEnabledLocationIdWithWebId(tx, webId)
			if err != nil {
				return nil, err
			}

			// 如果非Location
			if locationId == 0 {
				if !lists.ContainsInt64(result, webId) {
					result = append(result, webId)
				}
				break
			}

			// 查找包含此Location的Web
			// TODO 需要支持嵌套的Location查询
			webId, err = this.FindEnabledWebIdWithLocationId(tx, locationId)
			if err != nil {
				return nil, err
			}
			if webId == 0 {
				break
			}
		}
	}
	return result, nil
}

// 根据防火墙策略ID查找所有的WebId
func (this *HTTPWebDAO) FindAllWebIdsWithHTTPFirewallPolicyId(tx *dbs.Tx, firewallPolicyId int64) ([]int64, error) {
	ones, err := this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where(`JSON_CONTAINS(firewall, '{"isOn": true, "firewallPolicyId": ` + strconv.FormatInt(firewallPolicyId, 10) + ` }')`).
		Reuse(false). // 由于我们在JSON_CONTAINS()直接使用了变量，所以不能重用
		FindAll()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		webId := int64(one.(*HTTPWeb).Id)

		// 判断是否为Location
		for {
			locationId, err := SharedHTTPLocationDAO.FindEnabledLocationIdWithWebId(tx, webId)
			if err != nil {
				return nil, err
			}

			// 如果非Location
			if locationId == 0 {
				if !lists.ContainsInt64(result, webId) {
					result = append(result, webId)
				}
				break
			}

			// 查找包含此Location的Web
			// TODO 需要支持嵌套的Location查询
			webId, err = this.FindEnabledWebIdWithLocationId(tx, locationId)
			if err != nil {
				return nil, err
			}
			if webId == 0 {
				break
			}
		}
	}
	return result, nil
}

// 查找包含某个Location的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithLocationId(tx *dbs.Tx, locationId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where(`JSON_CONTAINS(locations, '{"locationId": ` + strconv.FormatInt(locationId, 10) + ` }')`).
		Reuse(false). // 由于我们在JSON_CONTAINS()直接使用了变量，所以不能重用
		FindInt64Col(0)
}

// 查找使用此Web的Server
func (this *HTTPWebDAO) FindWebServerId(tx *dbs.Tx, webId int64) (serverId int64, err error) {
	if webId <= 0 {
		return 0, nil
	}
	serverId, err = SharedServerDAO.FindEnabledServerIdWithWebId(tx, webId)
	if err != nil {
		return
	}
	if serverId > 0 {
		return
	}

	// web在Location中的情况
	locationId, err := SharedHTTPLocationDAO.FindEnabledLocationIdWithWebId(tx, webId)
	if err != nil {
		return 0, err
	}
	if locationId == 0 {
		return
	}
	webId, err = this.FindEnabledWebIdWithLocationId(tx, locationId)
	if err != nil {
		return
	}
	if webId <= 0 {
		return
	}

	// 第二轮查找
	return this.FindWebServerId(tx, webId)
}

// 检查用户权限
func (this *HTTPWebDAO) CheckUserWeb(tx *dbs.Tx, userId int64, webId int64) error {
	serverId, err := this.FindWebServerId(tx, webId)
	if err != nil {
		return err
	}
	if serverId == 0 {
		return ErrNotFound
	}
	return SharedServerDAO.CheckUserServer(tx, serverId, userId)
}

// 设置主机跳转
func (this *HTTPWebDAO) UpdateWebHostRedirects(tx *dbs.Tx, webId int64, hostRedirects []*serverconfigs.HTTPHostRedirectConfig) error {
	if webId <= 0 {
		return errors.New("invalid ")
	}
	if hostRedirects == nil {
		hostRedirects = []*serverconfigs.HTTPHostRedirectConfig{}
	}
	hostRedirectsJSON, err := json.Marshal(hostRedirects)
	if err != nil {
		return err
	}
	_, err = this.Query(tx).
		Pk(webId).
		Set("hostRedirects", hostRedirectsJSON).
		Update()
	return err
}

// 查找主机跳转
func (this *HTTPWebDAO) FindWebHostRedirects(tx *dbs.Tx, webId int64) ([]byte, error) {
	col, err := this.Query(tx).
		Pk(webId).
		Result("hostRedirects").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	return []byte(col), nil
}
