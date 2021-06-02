// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/maps"
)

type UserEdgeDNSProvider struct {
}

// Auth 认证
func (this *UserEdgeDNSProvider) Auth(params maps.Map) error {
	// TODO
	return nil
}

// GetRecords 获取域名解析记录列表
func (this *UserEdgeDNSProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	// TODO
	return
}

// GetRoutes 读取域名支持的线路数据
func (this *UserEdgeDNSProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	// TODO
	return
}

// QueryRecord 查询单个记录
func (this *UserEdgeDNSProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	// TODO
	return nil, nil
}

// AddRecord 设置记录
func (this *UserEdgeDNSProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	// TODO
	return nil
}

// UpdateRecord 修改记录
func (this *UserEdgeDNSProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	// TODO
	return nil
}

// DeleteRecord 删除记录
func (this *UserEdgeDNSProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	// TODO
	return nil
}

// DefaultRoute 默认线路
func (this *UserEdgeDNSProvider) DefaultRoute() string {
	// TODO
	return ""
}
