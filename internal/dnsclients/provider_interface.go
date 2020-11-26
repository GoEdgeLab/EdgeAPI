package dnsclients

import "github.com/iwind/TeaGo/maps"

// DNS操作接口
type ProviderInterface interface {
	// 认证
	Auth(params maps.Map) error

	// 获取域名解析记录列表
	GetRecords(domain string) (records []*Record, err error)

	// 读取域名支持的线路数据
	GetRoutes(domain string) (routes []*Route, err error)

	// 查询单个记录
	QueryRecord(domain string, name string, recordType RecordType) (*Record, error)

	// 设置记录
	AddRecord(domain string, newRecord *Record) error

	// 修改记录
	UpdateRecord(domain string, record *Record, newRecord *Record) error

	// 删除记录
	DeleteRecord(domain string, record *Record) error

	// 默认线路
	DefaultRoute() string
}
