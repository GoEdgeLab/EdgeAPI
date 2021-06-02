// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"regexp"
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
		result, err := nameservers.SharedNSRecordDAO.ListEnabledRecords(tx, domainId, "", "", 0, offset, size)
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
			var routeIdString = ""
			if len(routeIds) > 0 {
				routeIdString = fmt.Sprintf("%d", routeIds[0])
			}

			records = append(records, &dnstypes.Record{
				Id:    fmt.Sprintf("%d", record.Id),
				Name:  record.Name,
				Type:  record.Type,
				Value: record.Value,
				Route: routeIdString,
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

	// TODO 将来支持集群、域名、用户自定义线路
	result, err := nameservers.SharedNSRouteDAO.FindAllEnabledRoutes(tx, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	for _, route := range result {
		routes = append(routes, &dnstypes.Route{
			Name: route.Name,
			Code: fmt.Sprintf("%d", route.Id),
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
		routeIdString = fmt.Sprintf("%d", routeIds[0])
	}

	return &dnstypes.Record{
		Id:    fmt.Sprintf("%d", record.Id),
		Name:  record.Name,
		Type:  record.Type,
		Value: record.Value,
		Route: routeIdString,
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

	var routeIds []int64
	if len(newRecord.Route) > 0 && regexp.MustCompile(`^\d+$`).MatchString(newRecord.Route) {
		routeId := types.Int64(newRecord.Route)
		if routeId > 0 {
			routeIds = append(routeIds, routeId)
		}
	}

	_, err = nameservers.SharedNSRecordDAO.CreateRecord(tx, domainId, "", newRecord.Name, newRecord.Type, newRecord.Value, this.ttl, routeIds)
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

	var routeIds []int64
	if len(newRecord.Route) > 0 && regexp.MustCompile(`^\d+$`).MatchString(newRecord.Route) {
		routeId := types.Int64(newRecord.Route)
		if routeId > 0 {
			routeIds = append(routeIds, routeId)
		}
	}

	if len(record.Id) > 0 {
		err = nameservers.SharedNSRecordDAO.UpdateRecord(tx, types.Int64(record.Id), "", newRecord.Name, newRecord.Type, newRecord.Value, this.ttl, routeIds)
		if err != nil {
			return err
		}
	} else {
		realRecord, err := nameservers.SharedNSRecordDAO.FindEnabledRecordWithName(tx, domainId, record.Name, record.Type)
		if err != nil {
			return err
		}
		if realRecord != nil {
			err = nameservers.SharedNSRecordDAO.UpdateRecord(tx, types.Int64(realRecord.Id), "", newRecord.Name, newRecord.Type, newRecord.Value, this.ttl, routeIds)
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
	return ""
}
