package dnsclients

import (
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/maps"
)

// ProviderInterface DNS操作接口
type ProviderInterface interface {
	// Auth 认证
	Auth(params maps.Map) error

	// GetDomains 获取所有域名列表
	GetDomains() (domains []string, err error)

	// GetRecords 获取域名解析记录列表
	GetRecords(domain string) (records []*dnstypes.Record, err error)

	// GetRoutes 读取域名支持的线路数据
	GetRoutes(domain string) (routes []*dnstypes.Route, err error)

	// QueryRecord 查询单个记录
	QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error)

	// AddRecord 设置记录
	AddRecord(domain string, newRecord *dnstypes.Record) error

	// UpdateRecord 修改记录
	UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error

	// DeleteRecord 删除记录
	DeleteRecord(domain string, record *dnstypes.Record) error

	// DefaultRoute 默认线路
	DefaultRoute() string
}
