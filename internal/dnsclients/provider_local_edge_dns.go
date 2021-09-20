// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"strings"
)

type LocalEdgeDNSProvider struct {
	clusterId int64 // 集群ID
	ttl       int32 // TTL
}

// Auth 认证
func (this *LocalEdgeDNSProvider) Auth(params maps.Map) error {
	this.clusterId = params.GetInt64("clusterId")
	if this.clusterId <= 0 {
		return errors.New("'clusterId' should be greater than 0")
	}

	this.ttl = params.GetInt32("ttl")
	if this.ttl <= 0 {
		this.ttl = 3600
	}

	return nil
}

// GetDomains 获取所有域名列表
func (this *LocalEdgeDNSProvider) GetDomains() (domains []string, err error) {
	var tx *dbs.Tx
	domainOnes, err := nameservers.SharedNSDomainDAO.ListEnabledDomains(tx, this.clusterId, 0, "", 0, 1000)
	if err != nil {
		return nil, err
	}
	for _, domain := range domainOnes {
		domains = append(domains, domain.Name)
	}
	return
}

// GetRecords 获取域名解析记录列表
func (this *LocalEdgeDNSProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	var tx *dbs.Tx
	domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, this.clusterId, domain)
	if err != nil {
		return nil, err
	}
	if domainId == 0 {
		return nil, errors.New("can not find domain '" + domain + "'")
	}

	offset := int64(0)
	size := int64(1000)
	for {
		result, err := nameservers.SharedNSRecordDAO.ListEnabledRecords(tx, domainId, "", "", "", offset, size)
		if err != nil {
			return nil, err
		}
		if len(result) == 0 {
			break
		}
		for _, record := range result {
			if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Value, ".") {
				record.Value += "."
			}

			routeIds := record.DecodeRouteIds()
			if len(routeIds) == 0 {
				routeIds = []string{dnsconfigs.DefaultRouteCode}
			}
			records = append(records, &dnstypes.Record{
				Id:    fmt.Sprintf("%d", record.Id),
				Name:  record.Name,
				Type:  record.Type,
				Value: record.Value,
				Route: routeIds[0],
				TTL:   types.Int32(record.Ttl),
			})
		}

		offset += size
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *LocalEdgeDNSProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	var tx *dbs.Tx
	domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, this.clusterId, domain)
	if err != nil {
		return nil, err
	}
	if domainId == 0 {
		return nil, errors.New("can not find domain '" + domain + "'")
	}

	// 默认线路
	for _, route := range dnsconfigs.AllDefaultRoutes {
		routes = append(routes, &dnstypes.Route{
			Name: route.Name,
			Code: route.Code,
		})
	}

	// 自定义线路
	result, err := nameservers.SharedNSRouteDAO.FindAllEnabledRoutes(tx, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	for _, route := range result {
		routes = append(routes, &dnstypes.Route{
			Name: route.Name,
			Code: "id:" + types.String(route.Id),
		})
	}

	// 默认ISP
	for _, route := range dnsconfigs.AllDefaultISPRoutes {
		routes = append(routes, &dnstypes.Route{
			Name: route.Name,
			Code: route.Code,
		})
	}

	// 默认中国省份
	for _, route := range dnsconfigs.AllDefaultChinaProvinceRoutes {
		routes = append(routes, &dnstypes.Route{
			Name: route.Name,
			Code: route.Code,
		})
	}

	// 默认全球国家/地区
	for _, route := range dnsconfigs.AllDefaultWorldRegionRoutes {
		routes = append(routes, &dnstypes.Route{
			Name: route.Name,
			Code: route.Code,
		})
	}

	return
}

// QueryRecord 查询单个记录
func (this *LocalEdgeDNSProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	var tx *dbs.Tx
	domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, this.clusterId, domain)
	if err != nil {
		return nil, err
	}
	if domainId == 0 {
		return nil, nil
	}

	record, err := nameservers.SharedNSRecordDAO.FindEnabledRecordWithName(tx, domainId, name, recordType)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	routeIds := record.DecodeRouteIds()
	var routeIdString = ""
	if len(routeIds) > 0 {
		routeIdString = routeIds[0]
	} else {
		routeIdString = dnsconfigs.DefaultRouteCode
	}

	return &dnstypes.Record{
		Id:    fmt.Sprintf("%d", record.Id),
		Name:  record.Name,
		Type:  record.Type,
		Value: record.Value,
		Route: routeIdString,
		TTL:   types.Int32(record.Ttl),
	}, nil
}

// AddRecord 设置记录
func (this *LocalEdgeDNSProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	var tx *dbs.Tx
	domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, this.clusterId, domain)
	if err != nil {
		return err
	}
	if domainId == 0 {
		return errors.New("can not find domain '" + domain + "'")
	}

	var routeIds = []string{}
	if len(newRecord.Route) > 0 {
		routeIds = append(routeIds, newRecord.Route)
	}

	if newRecord.TTL <= 0 {
		newRecord.TTL = this.ttl
	}
	_, err = nameservers.SharedNSRecordDAO.CreateRecord(tx, domainId, "", newRecord.Name, newRecord.Type, newRecord.Value, newRecord.TTL, routeIds)
	if err != nil {
		return err
	}

	return nil
}

// UpdateRecord 修改记录
func (this *LocalEdgeDNSProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	var tx *dbs.Tx
	domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, this.clusterId, domain)
	if err != nil {
		return err
	}
	if domainId == 0 {
		return errors.New("can not find domain '" + domain + "'")
	}

	var routeIds []string
	if len(newRecord.Route) > 0 {
		routeIds = append(routeIds, newRecord.Route)
	}

	if newRecord.TTL <= 0 {
		newRecord.TTL = this.ttl
	}

	if len(record.Id) > 0 {
		err = nameservers.SharedNSRecordDAO.UpdateRecord(tx, types.Int64(record.Id), "", newRecord.Name, newRecord.Type, newRecord.Value, newRecord.TTL, routeIds, true)
		if err != nil {
			return err
		}
	} else {
		realRecord, err := nameservers.SharedNSRecordDAO.FindEnabledRecordWithName(tx, domainId, record.Name, record.Type)
		if err != nil {
			return err
		}
		if realRecord != nil {
			err = nameservers.SharedNSRecordDAO.UpdateRecord(tx, types.Int64(realRecord.Id), "", newRecord.Name, newRecord.Type, newRecord.Value, newRecord.TTL, routeIds, true)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteRecord 删除记录
func (this *LocalEdgeDNSProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	var tx *dbs.Tx
	domainId, err := nameservers.SharedNSDomainDAO.FindDomainIdWithName(tx, this.clusterId, domain)
	if err != nil {
		return err
	}
	if domainId == 0 {
		return errors.New("can not find domain '" + domain + "'")
	}

	if len(record.Id) > 0 {
		err = nameservers.SharedNSRecordDAO.DisableNSRecord(tx, types.Int64(record.Id))
		if err != nil {
			return err
		}
	} else {
		realRecord, err := nameservers.SharedNSRecordDAO.FindEnabledRecordWithName(tx, domainId, record.Name, record.Type)
		if err != nil {
			return err
		}
		if realRecord != nil {
			err = nameservers.SharedNSRecordDAO.DisableNSRecord(tx, types.Int64(realRecord.Id))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DefaultRoute 默认线路
func (this *LocalEdgeDNSProvider) DefaultRoute() string {
	return "default"
}
