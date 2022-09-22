package models

import "github.com/iwind/TeaGo/dbs"

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
	Referers           dbs.JSON `field:"referers"`           // 防盗链设置
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
	Referers           any // 防盗链设置
}

func NewHTTPWebOperator() *HTTPWebOperator {
	return &HTTPWebOperator{}
}
