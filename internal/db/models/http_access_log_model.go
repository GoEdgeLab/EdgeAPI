package models

import "github.com/iwind/TeaGo/dbs"

// HTTPAccessLog 访问日志
type HTTPAccessLog struct {
	Id                  uint64   `field:"id"`                  // ID
	ServerId            uint32   `field:"serverId"`            // 服务ID
	NodeId              uint32   `field:"nodeId"`              // 节点ID
	Status              uint32   `field:"status"`              // 状态码
	CreatedAt           uint64   `field:"createdAt"`           // 创建时间
	Content             dbs.JSON `field:"content"`             // 日志内容
	RequestId           string   `field:"requestId"`           // 请求ID
	FirewallPolicyId    uint32   `field:"firewallPolicyId"`    // WAF策略ID
	FirewallRuleGroupId uint32   `field:"firewallRuleGroupId"` // WAF分组ID
	FirewallRuleSetId   uint32   `field:"firewallRuleSetId"`   // WAF集ID
	FirewallRuleId      uint32   `field:"firewallRuleId"`      // WAF规则ID
	RemoteAddr          string   `field:"remoteAddr"`          // IP地址
	Domain              string   `field:"domain"`              // 域名
	RequestBody         []byte   `field:"requestBody"`         // 请求内容
	ResponseBody        []byte   `field:"responseBody"`        // 响应内容
}

type HTTPAccessLogOperator struct {
	Id                  interface{} // ID
	ServerId            interface{} // 服务ID
	NodeId              interface{} // 节点ID
	Status              interface{} // 状态码
	CreatedAt           interface{} // 创建时间
	Content             interface{} // 日志内容
	RequestId           interface{} // 请求ID
	FirewallPolicyId    interface{} // WAF策略ID
	FirewallRuleGroupId interface{} // WAF分组ID
	FirewallRuleSetId   interface{} // WAF集ID
	FirewallRuleId      interface{} // WAF规则ID
	RemoteAddr          interface{} // IP地址
	Domain              interface{} // 域名
	RequestBody         interface{} // 请求内容
	ResponseBody        interface{} // 响应内容
}

func NewHTTPAccessLogOperator() *HTTPAccessLogOperator {
	return &HTTPAccessLogOperator{}
}
