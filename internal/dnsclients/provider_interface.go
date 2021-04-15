package dnsclients

import "github.com/iwind/TeaGo/maps"

// ProviderInterface DNS操作接口
type ProviderInterface interface {
	// Auth 认证
	Auth(params maps.Map) error

	// GetRecords 获取域名解析记录列表
	GetRecords(domain string) (records []*Record, err error)

	// GetRoutes 读取域名支持的线路数据
	GetRoutes(domain string) (routes []*Route, err error)

	// QueryRecord 查询单个记录
	QueryRecord(domain string, name string, recordType RecordType) (*Record, error)

	// AddRecord 设置记录
	AddRecord(domain string, newRecord *Record) error

	// UpdateRecord 修改记录
	UpdateRecord(domain string, record *Record, newRecord *Record) error

	// DeleteRecord 删除记录
	DeleteRecord(domain string, record *Record) error

	// DefaultRoute 默认线路
	DefaultRoute() string
}
