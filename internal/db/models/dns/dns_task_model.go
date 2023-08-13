package dns

import "github.com/iwind/TeaGo/dbs"

const (
	DNSTaskField_Id         dbs.FieldName = "id"         // ID
	DNSTaskField_ClusterId  dbs.FieldName = "clusterId"  // 集群ID
	DNSTaskField_ServerId   dbs.FieldName = "serverId"   // 服务ID
	DNSTaskField_NodeId     dbs.FieldName = "nodeId"     // 节点ID
	DNSTaskField_DomainId   dbs.FieldName = "domainId"   // 域名ID
	DNSTaskField_RecordName dbs.FieldName = "recordName" // 记录名
	DNSTaskField_Type       dbs.FieldName = "type"       // 任务类型
	DNSTaskField_UpdatedAt  dbs.FieldName = "updatedAt"  // 更新时间
	DNSTaskField_IsDone     dbs.FieldName = "isDone"     // 是否已完成
	DNSTaskField_IsOk       dbs.FieldName = "isOk"       // 是否成功
	DNSTaskField_Error      dbs.FieldName = "error"      // 错误信息
	DNSTaskField_Version    dbs.FieldName = "version"    // 版本
	DNSTaskField_CountFails dbs.FieldName = "countFails" // 尝试失败次数
)

// DNSTask DNS更新任务
type DNSTask struct {
	Id         uint64 `field:"id"`         // ID
	ClusterId  uint32 `field:"clusterId"`  // 集群ID
	ServerId   uint32 `field:"serverId"`   // 服务ID
	NodeId     uint32 `field:"nodeId"`     // 节点ID
	DomainId   uint32 `field:"domainId"`   // 域名ID
	RecordName string `field:"recordName"` // 记录名
	Type       string `field:"type"`       // 任务类型
	UpdatedAt  uint64 `field:"updatedAt"`  // 更新时间
	IsDone     bool   `field:"isDone"`     // 是否已完成
	IsOk       bool   `field:"isOk"`       // 是否成功
	Error      string `field:"error"`      // 错误信息
	Version    uint64 `field:"version"`    // 版本
	CountFails uint32 `field:"countFails"` // 尝试失败次数
}

type DNSTaskOperator struct {
	Id         any // ID
	ClusterId  any // 集群ID
	ServerId   any // 服务ID
	NodeId     any // 节点ID
	DomainId   any // 域名ID
	RecordName any // 记录名
	Type       any // 任务类型
	UpdatedAt  any // 更新时间
	IsDone     any // 是否已完成
	IsOk       any // 是否成功
	Error      any // 错误信息
	Version    any // 版本
	CountFails any // 尝试失败次数
}

func NewDNSTaskOperator() *DNSTaskOperator {
	return &DNSTaskOperator{}
}
