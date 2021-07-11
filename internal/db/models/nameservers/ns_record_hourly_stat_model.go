package nameservers

// NSRecordHourlyStat NS记录统计
type NSRecordHourlyStat struct {
	Id            uint64 `field:"id"`            // ID
	ClusterId     uint32 `field:"clusterId"`     // 集群ID
	NodeId        uint32 `field:"nodeId"`        // 节点ID
	DomainId      uint32 `field:"domainId"`      // 域名ID
	RecordId      uint64 `field:"recordId"`      // 记录ID
	Day           string `field:"day"`           // YYYYMMDD
	Hour          string `field:"hour"`          // YYYYMMDDHH
	CountRequests uint32 `field:"countRequests"` // 请求数
	Bytes         uint64 `field:"bytes"`         // 流量
}

type NSRecordHourlyStatOperator struct {
	Id            interface{} // ID
	ClusterId     interface{} // 集群ID
	NodeId        interface{} // 节点ID
	DomainId      interface{} // 域名ID
	RecordId      interface{} // 记录ID
	Day           interface{} // YYYYMMDD
	Hour          interface{} // YYYYMMDDHH
	CountRequests interface{} // 请求数
	Bytes         interface{} // 流量
}

func NewNSRecordHourlyStatOperator() *NSRecordHourlyStatOperator {
	return &NSRecordHourlyStatOperator{}
}
