package models

import "github.com/iwind/TeaGo/dbs"

const (
	IPItemField_Id                            dbs.FieldName = "id"                            // ID
	IPItemField_ListId                        dbs.FieldName = "listId"                        // 所属名单ID
	IPItemField_Type                          dbs.FieldName = "type"                          // 类型
	IPItemField_IpFrom                        dbs.FieldName = "ipFrom"                        // 开始IP
	IPItemField_IpTo                          dbs.FieldName = "ipTo"                          // 结束IP
	IPItemField_IpFromLong                    dbs.FieldName = "ipFromLong"                    // 开始IP整型（弃用）
	IPItemField_IpToLong                      dbs.FieldName = "ipToLong"                      // 结束IP整型（弃用）
	IPItemField_Version                       dbs.FieldName = "version"                       // 版本
	IPItemField_CreatedAt                     dbs.FieldName = "createdAt"                     // 创建时间
	IPItemField_UpdatedAt                     dbs.FieldName = "updatedAt"                     // 修改时间
	IPItemField_Reason                        dbs.FieldName = "reason"                        // 加入说明
	IPItemField_EventLevel                    dbs.FieldName = "eventLevel"                    // 事件级别
	IPItemField_State                         dbs.FieldName = "state"                         // 状态
	IPItemField_ExpiredAt                     dbs.FieldName = "expiredAt"                     // 过期时间
	IPItemField_ServerId                      dbs.FieldName = "serverId"                      // 有效范围服务ID
	IPItemField_NodeId                        dbs.FieldName = "nodeId"                        // 有效范围节点ID
	IPItemField_SourceNodeId                  dbs.FieldName = "sourceNodeId"                  // 来源节点ID
	IPItemField_SourceServerId                dbs.FieldName = "sourceServerId"                // 来源服务ID
	IPItemField_SourceHTTPFirewallPolicyId    dbs.FieldName = "sourceHTTPFirewallPolicyId"    // 来源策略ID
	IPItemField_SourceHTTPFirewallRuleGroupId dbs.FieldName = "sourceHTTPFirewallRuleGroupId" // 来源规则集分组ID
	IPItemField_SourceHTTPFirewallRuleSetId   dbs.FieldName = "sourceHTTPFirewallRuleSetId"   // 来源规则集ID
	IPItemField_SourceUserId                  dbs.FieldName = "sourceUserId"                  // 用户ID
	IPItemField_IsRead                        dbs.FieldName = "isRead"                        // 是否已读
)

// IPItem IP
type IPItem struct {
	Id                            uint64 `field:"id"`                            // ID
	ListId                        uint32 `field:"listId"`                        // 所属名单ID
	Type                          string `field:"type"`                          // 类型
	IpFrom                        string `field:"ipFrom"`                        // 开始IP
	IpTo                          string `field:"ipTo"`                          // 结束IP
	IpFromLong                    uint64 `field:"ipFromLong"`                    // 开始IP整型（弃用）
	IpToLong                      uint64 `field:"ipToLong"`                      // 结束IP整型（弃用）
	Version                       uint64 `field:"version"`                       // 版本
	CreatedAt                     uint64 `field:"createdAt"`                     // 创建时间
	UpdatedAt                     uint64 `field:"updatedAt"`                     // 修改时间
	Reason                        string `field:"reason"`                        // 加入说明
	EventLevel                    string `field:"eventLevel"`                    // 事件级别
	State                         uint8  `field:"state"`                         // 状态
	ExpiredAt                     uint64 `field:"expiredAt"`                     // 过期时间
	ServerId                      uint32 `field:"serverId"`                      // 有效范围服务ID
	NodeId                        uint32 `field:"nodeId"`                        // 有效范围节点ID
	SourceNodeId                  uint32 `field:"sourceNodeId"`                  // 来源节点ID
	SourceServerId                uint32 `field:"sourceServerId"`                // 来源服务ID
	SourceHTTPFirewallPolicyId    uint32 `field:"sourceHTTPFirewallPolicyId"`    // 来源策略ID
	SourceHTTPFirewallRuleGroupId uint32 `field:"sourceHTTPFirewallRuleGroupId"` // 来源规则集分组ID
	SourceHTTPFirewallRuleSetId   uint32 `field:"sourceHTTPFirewallRuleSetId"`   // 来源规则集ID
	SourceUserId                  uint64 `field:"sourceUserId"`                  // 用户ID
	IsRead                        bool   `field:"isRead"`                        // 是否已读
}

type IPItemOperator struct {
	Id                            any // ID
	ListId                        any // 所属名单ID
	Type                          any // 类型
	IpFrom                        any // 开始IP
	IpTo                          any // 结束IP
	IpFromLong                    any // 开始IP整型（弃用）
	IpToLong                      any // 结束IP整型（弃用）
	Version                       any // 版本
	CreatedAt                     any // 创建时间
	UpdatedAt                     any // 修改时间
	Reason                        any // 加入说明
	EventLevel                    any // 事件级别
	State                         any // 状态
	ExpiredAt                     any // 过期时间
	ServerId                      any // 有效范围服务ID
	NodeId                        any // 有效范围节点ID
	SourceNodeId                  any // 来源节点ID
	SourceServerId                any // 来源服务ID
	SourceHTTPFirewallPolicyId    any // 来源策略ID
	SourceHTTPFirewallRuleGroupId any // 来源规则集分组ID
	SourceHTTPFirewallRuleSetId   any // 来源规则集ID
	SourceUserId                  any // 用户ID
	IsRead                        any // 是否已读
}

func NewIPItemOperator() *IPItemOperator {
	return &IPItemOperator{}
}
