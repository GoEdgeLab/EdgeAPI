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
	"github.com/iwind/TeaGo/maps"
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

var SharedHTTPWebDAO *HTTPWebDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPWebDAO = NewHTTPWebDAO()
	})
}

func (this *HTTPWebDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPWeb 启用条目
func (this *HTTPWebDAO) EnableHTTPWeb(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPWebStateEnabled).
		Update()
	return err
}

// DisableHTTPWeb 禁用条目
func (this *HTTPWebDAO) DisableHTTPWeb(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPWebStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPWeb 查找启用中的条目
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

// ComposeWebConfig 组合配置
func (this *HTTPWebDAO) ComposeWebConfig(tx *dbs.Tx, webId int64, cacheMap maps.Map) (*serverconfigs.HTTPWebConfig, error) {
	if cacheMap == nil {
		cacheMap = maps.Map{}
	}
	var cacheKey = this.Table + ":config:" + types.String(webId)
	var cache = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*serverconfigs.HTTPWebConfig), nil
	}

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

	// compression
	if IsNotNull(web.Compression) {
		compression := &serverconfigs.HTTPCompressionConfig{}
		err = json.Unmarshal([]byte(web.Compression), compression)
		if err != nil {
			return nil, err
		}
		config.Compression = compression

		// gzip
		if compression.GzipRef != nil && compression.GzipRef.Id > 0 {
			gzipConfig, err := SharedHTTPGzipDAO.ComposeGzipConfig(tx, compression.GzipRef.Id)
			if err != nil {
				return nil, err
			}
			compression.Gzip = gzipConfig
		}

		// brotli
		if compression.BrotliRef != nil && compression.BrotliRef.Id > 0 {
			brotliConfig, err := SharedHTTPBrotliPolicyDAO.ComposeBrotliConfig(tx, compression.BrotliRef.Id)
			if err != nil {
				return nil, err
			}
			compression.Brotli = brotliConfig
		}

		// deflate
		if compression.DeflateRef != nil && compression.DeflateRef.Id > 0 {
			deflateConfig, err := SharedHTTPDeflatePolicyDAO.ComposeDeflateConfig(tx, compression.DeflateRef.Id)
			if err != nil {
				return nil, err
			}
			compression.Deflate = deflateConfig
		}
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
			pageConfig, err := SharedHTTPPageDAO.ComposePageConfig(tx, page.Id, cacheMap)
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
			firewallPolicy, err := SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, firewallRef.FirewallPolicyId, cacheMap)
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

	// 路由规则
	if IsNotNull(web.Locations) {
		refs := []*serverconfigs.HTTPLocationRef{}
		err = json.Unmarshal([]byte(web.Locations), &refs)
		if err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			config.LocationRefs = refs

			locations, err := SharedHTTPLocationDAO.ConvertLocationRefs(tx, refs, cacheMap)
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
			rewriteRule, err := SharedHTTPRewriteRuleDAO.ComposeRewriteRule(tx, ref.RewriteRuleId, cacheMap)
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

	// Fastcgi
	if IsNotNull(web.Fastcgi) {
		ref := &serverconfigs.HTTPFastcgiRef{}
		err = json.Unmarshal([]byte(web.Fastcgi), ref)
		if err != nil {
			return nil, err
		}
		config.FastcgiRef = ref

		if len(ref.FastcgiIds) > 0 {
			list := []*serverconfigs.HTTPFastcgiConfig{}
			for _, fastcgiId := range ref.FastcgiIds {
				fastcgiConfig, err := SharedHTTPFastcgiDAO.ComposeFastcgiConfig(tx, fastcgiId)
				if err != nil {
					return nil, err
				}
				if fastcgiConfig != nil {
					list = append(list, fastcgiConfig)
				}
			}
			config.FastcgiList = list
		}
	}

	// 认证
	if IsNotNull(web.Auth) {
		authConfig := &serverconfigs.HTTPAuthConfig{}
		err = json.Unmarshal([]byte(web.Auth), authConfig)
		if err != nil {
			return nil, err
		}
		var newRefs []*serverconfigs.HTTPAuthPolicyRef
		for _, ref := range authConfig.PolicyRefs {
			policyConfig, err := SharedHTTPAuthPolicyDAO.ComposePolicyConfig(tx, ref.AuthPolicyId, cacheMap)
			if err != nil {
				return nil, err
			}
			if policyConfig != nil {
				ref.AuthPolicy = policyConfig
				newRefs = append(newRefs, ref)
			}
		}
		config.Auth = authConfig
	}

	// WebP
	if IsNotNull(web.Webp) {
		var webpConfig = &serverconfigs.WebPImageConfig{}
		err = json.Unmarshal([]byte(web.Webp), webpConfig)
		if err != nil {
			return nil, err
		}
		config.WebP = webpConfig
	}

	// RemoteAddr
	if IsNotNull(web.RemoteAddr) {
		var remoteAddrConfig = &serverconfigs.HTTPRemoteAddrConfig{}
		err = json.Unmarshal([]byte(web.RemoteAddr), remoteAddrConfig)
		if err != nil {
			return nil, err
		}
		config.RemoteAddr = remoteAddrConfig
	}

	cacheMap[cacheKey] = config

	return config, nil
}

// CreateWeb 创建Web配置
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

// UpdateWeb 修改Web配置
func (this *HTTPWebDAO) UpdateWeb(tx *dbs.Tx, webId int64, rootJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Root = JSONBytes(rootJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebCompression 修改压缩配置
func (this *HTTPWebDAO) UpdateWebCompression(tx *dbs.Tx, webId int64, compressionConfig []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Compression = JSONBytes(compressionConfig)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebWebP 修改WebP配置
func (this *HTTPWebDAO) UpdateWebWebP(tx *dbs.Tx, webId int64, webpConfig []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Webp = JSONBytes(webpConfig)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebRemoteAddr 修改RemoteAddr配置
func (this *HTTPWebDAO) UpdateWebRemoteAddr(tx *dbs.Tx, webId int64, remoteAddrConfig []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	var op = NewHTTPWebOperator()
	op.Id = webId
	op.RemoteAddr = remoteAddrConfig
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, webId)
}

// UpdateWebCharset 修改字符编码
func (this *HTTPWebDAO) UpdateWebCharset(tx *dbs.Tx, webId int64, charsetJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Charset = JSONBytes(charsetJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebRequestHeaderPolicy 更改请求Header策略
func (this *HTTPWebDAO) UpdateWebRequestHeaderPolicy(tx *dbs.Tx, webId int64, headerPolicyJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.RequestHeader = JSONBytes(headerPolicyJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebResponseHeaderPolicy 更改响应Header策略
func (this *HTTPWebDAO) UpdateWebResponseHeaderPolicy(tx *dbs.Tx, webId int64, headerPolicyJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.ResponseHeader = JSONBytes(headerPolicyJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebPages 更改特殊页面配置
func (this *HTTPWebDAO) UpdateWebPages(tx *dbs.Tx, webId int64, pagesJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Pages = JSONBytes(pagesJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebShutdown 更改Shutdown配置
func (this *HTTPWebDAO) UpdateWebShutdown(tx *dbs.Tx, webId int64, shutdownJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Shutdown = JSONBytes(shutdownJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebAccessLogConfig 更改访问日志策略
func (this *HTTPWebDAO) UpdateWebAccessLogConfig(tx *dbs.Tx, webId int64, accessLogJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.AccessLog = JSONBytes(accessLogJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebStat 更改统计配置
func (this *HTTPWebDAO) UpdateWebStat(tx *dbs.Tx, webId int64, statJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Stat = JSONBytes(statJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebCache 更改缓存配置
func (this *HTTPWebDAO) UpdateWebCache(tx *dbs.Tx, webId int64, cacheJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Cache = JSONBytes(cacheJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebFirewall 更改防火墙配置
func (this *HTTPWebDAO) UpdateWebFirewall(tx *dbs.Tx, webId int64, firewallJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Firewall = JSONBytes(firewallJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebLocations 更改路由规则配置
func (this *HTTPWebDAO) UpdateWebLocations(tx *dbs.Tx, webId int64, locationsJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Locations = JSONBytes(locationsJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebRedirectToHTTPS 更改跳转到HTTPS设置
func (this *HTTPWebDAO) UpdateWebRedirectToHTTPS(tx *dbs.Tx, webId int64, redirectToHTTPSJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.RedirectToHttps = JSONBytes(redirectToHTTPSJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebsocket 修改Websocket设置
func (this *HTTPWebDAO) UpdateWebsocket(tx *dbs.Tx, webId int64, websocketJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Websocket = JSONBytes(websocketJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebFastcgi 修改Fastcgi设置
func (this *HTTPWebDAO) UpdateWebFastcgi(tx *dbs.Tx, webId int64, fastcgiJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Fastcgi = JSONBytes(fastcgiJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebRewriteRules 修改重写规则设置
func (this *HTTPWebDAO) UpdateWebRewriteRules(tx *dbs.Tx, webId int64, rewriteRulesJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.RewriteRules = JSONBytes(rewriteRulesJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// UpdateWebAuth 修改认证信息
func (this *HTTPWebDAO) UpdateWebAuth(tx *dbs.Tx, webId int64, authJSON []byte) error {
	if webId <= 0 {
		return errors.New("invalid webId")
	}
	op := NewHTTPWebOperator()
	op.Id = webId
	op.Auth = JSONBytes(authJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// FindAllWebIdsWithCachePolicyId 根据缓存策略ID查找所有的WebId
func (this *HTTPWebDAO) FindAllWebIdsWithCachePolicyId(tx *dbs.Tx, cachePolicyId int64) ([]int64, error) {
	ones, err := this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where(`JSON_CONTAINS(cache, :jsonQuery, '$.cacheRefs')`).
		Param("jsonQuery", maps.Map{"cachePolicyId": cachePolicyId}.AsJSON()).
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

// FindAllWebIdsWithHTTPFirewallPolicyId 根据防火墙策略ID查找所有的WebId
func (this *HTTPWebDAO) FindAllWebIdsWithHTTPFirewallPolicyId(tx *dbs.Tx, firewallPolicyId int64) ([]int64, error) {
	ones, err := this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where(`JSON_CONTAINS(firewall, :jsonQuery)`).
		Param("jsonQuery", maps.Map{
			// 这里不加入isOn的判断，无论是否开启我们都同步
			"firewallPolicyId": firewallPolicyId,
		}.AsJSON()).
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

// FindEnabledWebIdWithLocationId 查找包含某个Location的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithLocationId(tx *dbs.Tx, locationId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(locations, :jsonQuery)").
		Param("jsonQuery", maps.Map{"locationId": locationId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithRewriteRuleId 查找包含某个重写规则的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithRewriteRuleId(tx *dbs.Tx, rewriteRuleId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(rewriteRules, :jsonQuery)").
		Param("jsonQuery", maps.Map{"rewriteRuleId": rewriteRuleId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithPageId 查找包含某个页面的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithPageId(tx *dbs.Tx, pageId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(pages, :jsonQuery)").
		Param("jsonQuery", maps.Map{"id": pageId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithHeaderPolicyId 查找包含某个Header的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithHeaderPolicyId(tx *dbs.Tx, headerPolicyId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("(JSON_CONTAINS(requestHeader, :jsonQuery) OR JSON_CONTAINS(responseHeader, :jsonQuery))").
		Param("jsonQuery", maps.Map{"headerPolicyId": headerPolicyId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithGzipId 查找包含某个Gzip配置的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithGzipId(tx *dbs.Tx, gzipId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(compression, :jsonQuery, '$.gzipRef')").
		Param("jsonQuery", maps.Map{"id": gzipId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithBrotliPolicyId 查找包含某个Brotli配置的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithBrotliPolicyId(tx *dbs.Tx, brotliPolicyId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(compression, :jsonQuery, '$.brotliRef')").
		Param("jsonQuery", maps.Map{"id": brotliPolicyId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithDeflatePolicyId 查找包含某个Deflate配置的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithDeflatePolicyId(tx *dbs.Tx, deflatePolicyId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(compression, :jsonQuery, '$.deflateRef')").
		Param("jsonQuery", maps.Map{"id": deflatePolicyId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithWebsocketId 查找包含某个Websocket配置的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithWebsocketId(tx *dbs.Tx, websocketId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(websocket, :jsonQuery)").
		Param("jsonQuery", maps.Map{"websocketId": websocketId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithFastcgiId 查找包含某个Fastcgi配置的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithFastcgiId(tx *dbs.Tx, fastcgiId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(fastcgi, :jsonQuery)").
		Param("jsonQuery", maps.Map{"fastcgiIds": fastcgiId}.AsJSON()).
		FindInt64Col(0)
}

// FindEnabledWebIdWithHTTPAuthPolicyId 查找包含某个认证策略的Web
func (this *HTTPWebDAO) FindEnabledWebIdWithHTTPAuthPolicyId(tx *dbs.Tx, httpAuthPolicyId int64) (webId int64, err error) {
	return this.Query(tx).
		State(HTTPWebStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(auth, :jsonQuery, '$.policyRefs')").
		Param("jsonQuery", maps.Map{"authPolicyId": httpAuthPolicyId}.AsJSON()).
		FindInt64Col(0)
}

// FindWebServerId 查找使用此Web的Server
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

// CheckUserWeb 检查用户权限
func (this *HTTPWebDAO) CheckUserWeb(tx *dbs.Tx, userId int64, webId int64) error {
	serverId, err := this.FindWebServerId(tx, webId)
	if err != nil {
		return err
	}
	if serverId == 0 {
		return ErrNotFound
	}
	return SharedServerDAO.CheckUserServer(tx, userId, serverId)
}

// UpdateWebHostRedirects 设置主机跳转
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
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, webId)
}

// FindWebHostRedirects 查找主机跳转
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

// NotifyUpdate 通知更新
func (this *HTTPWebDAO) NotifyUpdate(tx *dbs.Tx, webId int64) error {
	serverId, err := this.FindWebServerId(tx, webId)
	if err != nil {
		return err
	}
	if serverId == 0 {
		return nil
	}
	return SharedServerDAO.NotifyUpdate(tx, serverId)
}
