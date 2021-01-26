package stats

// WAF统计
type ServerHTTPFirewallDailyStat struct {
	Id                      uint64 `field:"id"`                      // ID
	ServerId                uint32 `field:"serverId"`                // 服务ID
	Day                     string `field:"day"`                     // YYYYMMDD
	HttpFirewallRuleGroupId uint32 `field:"httpFirewallRuleGroupId"` // WAF分组ID
	Action                  string `field:"action"`                  // 采取的动作
	Count                   uint64 `field:"count"`                   // 数量
}

type ServerHTTPFirewallDailyStatOperator struct {
	Id                      interface{} // ID
	ServerId                interface{} // 服务ID
	Day                     interface{} // YYYYMMDD
	HttpFirewallRuleGroupId interface{} // WAF分组ID
	Action                  interface{} // 采取的动作
	Count                   interface{} // 数量
}

func NewServerHTTPFirewallDailyStatOperator() *ServerHTTPFirewallDailyStatOperator {
	return &ServerHTTPFirewallDailyStatOperator{}
}
