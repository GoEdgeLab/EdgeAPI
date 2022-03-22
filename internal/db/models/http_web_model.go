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
}

type HTTPWebOperator struct {
	Id                 interface{} // ID
	IsOn               interface{} // 是否启用
	TemplateId         interface{} // 模版ID
	AdminId            interface{} // 管理员ID
	UserId             interface{} // 用户ID
	State              interface{} // 状态
	CreatedAt          interface{} // 创建时间
	Root               interface{} // 根目录
	Charset            interface{} // 字符集
	Shutdown           interface{} // 临时关闭页面配置
	Pages              interface{} // 特殊页面
	RedirectToHttps    interface{} // 跳转到HTTPS设置
	Indexes            interface{} // 首页文件列表
	MaxRequestBodySize interface{} // 最大允许的请求内容尺寸
	RequestHeader      interface{} // 请求Header配置
	ResponseHeader     interface{} // 响应Header配置
	AccessLog          interface{} // 访问日志配置
	Stat               interface{} // 统计配置
	Gzip               interface{} // Gzip配置（v0.3.2弃用）
	Compression        interface{} // 压缩配置
	Cache              interface{} // 缓存配置
	Firewall           interface{} // 防火墙设置
	Locations          interface{} // 路由规则配置
	Websocket          interface{} // Websocket设置
	RewriteRules       interface{} // 重写规则配置
	HostRedirects      interface{} // 域名跳转
	Fastcgi            interface{} // Fastcgi配置
	Auth               interface{} // 认证策略配置
	Webp               interface{} // WebP配置
	RemoteAddr         interface{} // 客户端IP配置
	MergeSlashes       interface{} // 是否合并路径中的斜杠
	RequestLimit       interface{} // 请求限制
	RequestScripts     interface{} // 请求脚本
}

func NewHTTPWebOperator() *HTTPWebOperator {
	return &HTTPWebOperator{}
}
