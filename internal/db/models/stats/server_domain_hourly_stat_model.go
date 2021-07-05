package stats

// ServerDomainHourlyStat 服务域名统计
type ServerDomainHourlyStat struct {
	Id                  uint64 `field:"id"`                  // ID
	ClusterId           uint32 `field:"clusterId"`           // 集群ID
	NodeId              uint32 `field:"nodeId"`              // 节点ID
	ServerId            uint32 `field:"serverId"`            // 服务ID
	Domain              string `field:"domain"`              // 域名
	Hour                string `field:"hour"`                // YYYYMMDDHH
	Bytes               uint64 `field:"bytes"`               // 流量
	CachedBytes         uint64 `field:"cachedBytes"`         // 缓存流量
	CountRequests       uint64 `field:"countRequests"`       // 请求数
	CountCachedRequests uint64 `field:"countCachedRequests"` // 缓存请求
}

type ServerDomainHourlyStatOperator struct {
	Id                  interface{} // ID
	ClusterId           interface{} // 集群ID
	NodeId              interface{} // 节点ID
	ServerId            interface{} // 服务ID
	Domain              interface{} // 域名
	Hour                interface{} // YYYYMMDDHH
	Bytes               interface{} // 流量
	CachedBytes         interface{} // 缓存流量
	CountRequests       interface{} // 请求数
	CountCachedRequests interface{} // 缓存请求
}

func NewServerDomainHourlyStatOperator() *ServerDomainHourlyStatOperator {
	return &ServerDomainHourlyStatOperator{}
}
