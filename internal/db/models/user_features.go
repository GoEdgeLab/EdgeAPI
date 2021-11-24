package models

import "github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"

var (
	// 所有功能列表，注意千万不能在运行时进行修改
	allUserFeatures = []*UserFeature{
		{
			Name:        "记录访问日志",
			Code:        "server.accessLog",
			Description: "用户可以开启服务的访问日志",
		},
		{
			Name:        "查看访问日志",
			Code:        "server.viewAccessLog",
			Description: "用户可以查看服务的访问日志",
		},
		{
			Name:        "转发访问日志",
			Code:        "server.accessLog.forward",
			Description: "用户可以配置访问日志转发到自定义的API",
		},
		{
			Name:        "TCP负载均衡",
			Code:        "server.tcp",
			Description: "用户可以添加TCP/TLS负载均衡服务",
		},
		{
			Name:        "自定义TCP负载均衡端口",
			Code:        "server.tcp.port",
			Description: "用户可以自定义TCP端口",
		},
		{
			Name:        "UDP负载均衡",
			Code:        "server.udp",
			Description: "用户可以添加UDP负载均衡服务",
		},
		{
			Name:        "自定义UDP负载均衡端口",
			Code:        "server.udp.port",
			Description: "用户可以自定义UDP端口",
		},
		{
			Name:        "开启WAF",
			Code:        "server.waf",
			Description: "用户可以开启WAF功能并可以设置黑白名单等",
		},
		{
			Name:        "费用账单",
			Code:        "finance",
			Description: "开启费用账单相关功能",
		},
	}
)

// UserFeature 用户功能
type UserFeature struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (this *UserFeature) ToPB() *pb.UserFeature {
	return &pb.UserFeature{Name: this.Name, Code: this.Code, Description: this.Description}
}

// FindAllUserFeatures 所有功能列表
func FindAllUserFeatures() []*UserFeature {
	return allUserFeatures
}

// FindUserFeature 查询单个功能
func FindUserFeature(code string) *UserFeature {
	for _, feature := range allUserFeatures {
		if feature.Code == code {
			return feature
		}
	}
	return nil
}
