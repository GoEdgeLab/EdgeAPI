package models

import "github.com/iwind/TeaGo/dbs"

const (
	HTTPWebField_Id                 dbs.FieldName = "id"                 // ID
	HTTPWebField_IsOn               dbs.FieldName = "isOn"               // 是否启用
	HTTPWebField_TemplateId         dbs.FieldName = "templateId"         // 模版ID
	HTTPWebField_AdminId            dbs.FieldName = "adminId"            // 管理员ID
	HTTPWebField_UserId             dbs.FieldName = "userId"             // 用户ID
	HTTPWebField_State              dbs.FieldName = "state"              // 状态
	HTTPWebField_CreatedAt          dbs.FieldName = "createdAt"          // 创建时间
	HTTPWebField_Root               dbs.FieldName = "root"               // 根目录
	HTTPWebField_Charset            dbs.FieldName = "charset"            // 字符集
	HTTPWebField_Shutdown           dbs.FieldName = "shutdown"           // 临时关闭页面配置
	HTTPWebField_Pages              dbs.FieldName = "pages"              // 特殊页面
	HTTPWebField_RedirectToHttps    dbs.FieldName = "redirectToHttps"    // 跳转到HTTPS设置
	HTTPWebField_Indexes            dbs.FieldName = "indexes"            // 首页文件列表
	HTTPWebField_MaxRequestBodySize dbs.FieldName = "maxRequestBodySize" // 最大允许的请求内容尺寸
	HTTPWebField_RequestHeader      dbs.FieldName = "requestHeader"      // 请求Header配置
	HTTPWebField_ResponseHeader     dbs.FieldName = "responseHeader"     // 响应Header配置
	HTTPWebField_AccessLog          dbs.FieldName = "accessLog"          // 访问日志配置
	HTTPWebField_Stat               dbs.FieldName = "stat"               // 统计配置
	HTTPWebField_Gzip               dbs.FieldName = "gzip"               // Gzip配置（v0.3.2弃用）
	HTTPWebField_Compression        dbs.FieldName = "compression"        // 压缩配置
	HTTPWebField_Cache              dbs.FieldName = "cache"              // 缓存配置
	HTTPWebField_Firewall           dbs.FieldName = "firewall"           // 防火墙设置
	HTTPWebField_Locations          dbs.FieldName = "locations"          // 路由规则配置
	HTTPWebField_Websocket          dbs.FieldName = "websocket"          // Websocket设置
	HTTPWebField_RewriteRules       dbs.FieldName = "rewriteRules"       // 重写规则配置
	HTTPWebField_HostRedirects      dbs.FieldName = "hostRedirects"      // 域名跳转
	HTTPWebField_Fastcgi            dbs.FieldName = "fastcgi"            // Fastcgi配置
	HTTPWebField_Auth               dbs.FieldName = "auth"               // 认证策略配置
	HTTPWebField_Webp               dbs.FieldName = "webp"               // WebP配置
	HTTPWebField_RemoteAddr         dbs.FieldName = "remoteAddr"         // 客户端IP配置
	HTTPWebField_MergeSlashes       dbs.FieldName = "mergeSlashes"       // 是否合并路径中的斜杠
	HTTPWebField_RequestLimit       dbs.FieldName = "requestLimit"       // 请求限制
	HTTPWebField_RequestScripts     dbs.FieldName = "requestScripts"     // 请求脚本
	HTTPWebField_Uam                dbs.FieldName = "uam"                // UAM设置
	HTTPWebField_Cc                 dbs.FieldName = "cc"                 // CC设置
	HTTPWebField_Referers           dbs.FieldName = "referers"           // 防盗链设置
	HTTPWebField_UserAgent          dbs.FieldName = "userAgent"          // UserAgent设置
	HTTPWebField_Optimization       dbs.FieldName = "optimization"       // 页面优化配置
)

// HTTPWeb HTTP Web
type HTTPWeb struct {
	Id                 uint32   `field:"id"`                 // ID
	IsOn               bool     `field:"isOn"`               // 是否启用
	TemplateId         uint32   `field:"templateId"`         // 模版ID
	AdminId            uint32   `field:"adminId"`            // 管理员ID
	UserId             uint32   `field:"userId"`             // 用户ID
	State              uint8    `field:"state"`              // 状态
	CreatedAt          uint64   `field:"createdAt"`          // 创建时间
	Root               dbs.JSON `field:"root"`               // 根目录
	Charset            dbs.JSON `field:"charset"`            // 字符集
	Shutdown           dbs.JSON `field:"shutdown"`           // 临时关闭页面配置
	Pages              dbs.JSON `field:"pages"`              // 特殊页面
	RedirectToHttps    dbs.JSON `field:"redirectToHttps"`    // 跳转到HTTPS设置
	Indexes            dbs.JSON `field:"indexes"`            // 首页文件列表
	MaxRequestBodySize dbs.JSON `field:"maxRequestBodySize"` // 最大允许的请求内容尺寸
	RequestHeader      dbs.JSON `field:"requestHeader"`      // 请求Header配置
	ResponseHeader     dbs.JSON `field:"responseHeader"`     // 响应Header配置
	AccessLog          dbs.JSON `field:"accessLog"`          // 访问日志配置
	Stat               dbs.JSON `field:"stat"`               // 统计配置
	Gzip               dbs.JSON `field:"gzip"`               // Gzip配置（v0.3.2弃用）
	Compression        dbs.JSON `field:"compression"`        // 压缩配置
	Cache              dbs.JSON `field:"cache"`              // 缓存配置
	Firewall           dbs.JSON `field:"firewall"`           // 防火墙设置
	Locations          dbs.JSON `field:"locations"`          // 路由规则配置
	Websocket          dbs.JSON `field:"websocket"`          // Websocket设置
	RewriteRules       dbs.JSON `field:"rewriteRules"`       // 重写规则配置
	HostRedirects      dbs.JSON `field:"hostRedirects"`      // 域名跳转
	Fastcgi            dbs.JSON `field:"fastcgi"`            // Fastcgi配置
	Auth               dbs.JSON `field:"auth"`               // 认证策略配置
	Webp               dbs.JSON `field:"webp"`               // WebP配置
	RemoteAddr         dbs.JSON `field:"remoteAddr"`         // 客户端IP配置
	MergeSlashes       uint8    `field:"mergeSlashes"`       // 是否合并路径中的斜杠
	RequestLimit       dbs.JSON `field:"requestLimit"`       // 请求限制
	RequestScripts     dbs.JSON `field:"requestScripts"`     // 请求脚本
	Uam                dbs.JSON `field:"uam"`                // UAM设置
	Cc                 dbs.JSON `field:"cc"`                 // CC设置
	Referers           dbs.JSON `field:"referers"`           // 防盗链设置
	UserAgent          dbs.JSON `field:"userAgent"`          // UserAgent设置
	Optimization       dbs.JSON `field:"optimization"`       // 页面优化配置
}

type HTTPWebOperator struct {
	Id                 any // ID
	IsOn               any // 是否启用
	TemplateId         any // 模版ID
	AdminId            any // 管理员ID
	UserId             any // 用户ID
	State              any // 状态
	CreatedAt          any // 创建时间
	Root               any // 根目录
	Charset            any // 字符集
	Shutdown           any // 临时关闭页面配置
	Pages              any // 特殊页面
	RedirectToHttps    any // 跳转到HTTPS设置
	Indexes            any // 首页文件列表
	MaxRequestBodySize any // 最大允许的请求内容尺寸
	RequestHeader      any // 请求Header配置
	ResponseHeader     any // 响应Header配置
	AccessLog          any // 访问日志配置
	Stat               any // 统计配置
	Gzip               any // Gzip配置（v0.3.2弃用）
	Compression        any // 压缩配置
	Cache              any // 缓存配置
	Firewall           any // 防火墙设置
	Locations          any // 路由规则配置
	Websocket          any // Websocket设置
	RewriteRules       any // 重写规则配置
	HostRedirects      any // 域名跳转
	Fastcgi            any // Fastcgi配置
	Auth               any // 认证策略配置
	Webp               any // WebP配置
	RemoteAddr         any // 客户端IP配置
	MergeSlashes       any // 是否合并路径中的斜杠
	RequestLimit       any // 请求限制
	RequestScripts     any // 请求脚本
	Uam                any // UAM设置
	Cc                 any // CC设置
	Referers           any // 防盗链设置
	UserAgent          any // UserAgent设置
	Optimization       any // 页面优化配置
}

func NewHTTPWebOperator() *HTTPWebOperator {
	return &HTTPWebOperator{}
}
