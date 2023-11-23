// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsclients

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tencenterrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"strings"
)

// TencentDNSProvider 腾讯云DNS云解析
type TencentDNSProvider struct {
	BaseProvider

	ProviderId int64

	client *dnspod.Client
}

func NewTencentDNSProvider() *TencentDNSProvider {
	return &TencentDNSProvider{}
}

// Auth 认证
func (this *TencentDNSProvider) Auth(params maps.Map) error {
	var accessKeyId = params.GetString("accessKeyId")
	var accessKeySecret = params.GetString("accessKeySecret")
	if len(accessKeyId) == 0 {
		return errors.New("'accessKeyId' required")
	}
	if len(accessKeySecret) == 0 {
		return errors.New("'accessKeySecret' required")
	}

	client, err := dnspod.NewClient(common.NewCredential(accessKeyId, accessKeySecret), "", profile.NewClientProfile())
	if err != nil {
		return err
	}
	this.client = client

	return nil
}

// GetDomains 获取所有域名列表
func (this *TencentDNSProvider) GetDomains() (domains []string, err error) {
	var offset int64 = 0
	var limit int64 = 1000
	for {
		var req = dnspod.NewDescribeDomainListRequest()
		req.Offset = this.int64Val(offset)
		req.Limit = this.int64Val(limit)
		resp, respErr := this.client.DescribeDomainList(req)
		if respErr != nil {
			if this.isNotFoundErr(respErr) {
				break
			}
			return nil, respErr
		}
		var countDomains = len(resp.Response.DomainList)
		if countDomains == 0 {
			break
		}
		for _, domainObj := range resp.Response.DomainList {
			domains = append(domains, *domainObj.Name)
		}
		offset += int64(countDomains)
	}

	return
}

// GetRecords 获取域名列表
func (this *TencentDNSProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	var offset uint64 = 0
	var limit uint64 = 1000
	for {
		var req = dnspod.NewDescribeRecordListRequest()
		req.Domain = this.stringVal(domain)
		req.Offset = this.uint64Val(offset)
		req.Limit = this.uint64Val(limit)
		resp, respErr := this.client.DescribeRecordList(req)
		if respErr != nil {
			if this.isNotFoundErr(respErr) {
				break
			}
			return nil, respErr
		}
		var countRecords = len(resp.Response.RecordList)
		if countRecords == 0 {
			break
		}
		for _, recordObj := range resp.Response.RecordList {
			records = append(records, &dnstypes.Record{
				Id:    types.String(*recordObj.RecordId),
				Name:  *recordObj.Name,
				Type:  *recordObj.Type,
				Value: this.fixCNAME(*recordObj.Type, *recordObj.Value),
				Route: *recordObj.LineId,
				TTL:   types.Int32(*recordObj.TTL),
			})
		}
		offset += uint64(countRecords)
	}

	// 写入缓存
	if this.ProviderId > 0 {
		sharedDomainRecordsCache.WriteDomainRecords(this.ProviderId, domain, records)
	}

	return
}

// GetRoutes 读取线路数据
func (this *TencentDNSProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	// 等级信息
	var domainGrade string
	{
		var req = dnspod.NewDescribeDomainRequest()
		req.Domain = this.stringVal(domain)
		resp, respErr := this.client.DescribeDomain(req)
		if respErr != nil {
			if this.isNotFoundErr(respErr) {
				return
			}
			return nil, respErr
		}
		if resp.Response.DomainInfo == nil {
			return
		}
		domainGrade = *resp.Response.DomainInfo.Grade
	}

	// 等级允许的线路
	{
		var req = dnspod.NewDescribeRecordLineListRequest()
		req.Domain = this.stringVal(domain)
		req.DomainGrade = this.stringVal(domainGrade)
		resp, respErr := this.client.DescribeRecordLineList(req)
		if respErr != nil {
			return nil, respErr
		}
		for _, lineGroupObj := range resp.Response.LineGroupList {
			routes = append(routes, &dnstypes.Route{
				Name: "Group:" + *lineGroupObj.Name,
				Code: *lineGroupObj.LineId,
			})
		}
		for _, lineObj := range resp.Response.LineList {
			routes = append(routes, &dnstypes.Route{
				Name: *lineObj.Name,
				Code: *lineObj.LineId,
			})
		}
	}

	return
}

// QueryRecord 查询单个记录
func (this *TencentDNSProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	// 从缓存中读取
	if this.ProviderId > 0 {
		record, hasRecords, _ := sharedDomainRecordsCache.QueryDomainRecord(this.ProviderId, domain, name, recordType)
		if hasRecords { // 有效的搜索
			return record, nil
		}
	}

	var offset uint64 = 0
	var limit uint64 = 1000
	var req = dnspod.NewDescribeRecordFilterListRequest()
	req.Domain = this.stringVal(domain)
	req.Offset = this.uint64Val(offset)
	req.Limit = this.uint64Val(limit)
	req.SubDomain = this.stringVal(name)
	req.RecordType = []*string{this.stringVal(recordType)}
	resp, respErr := this.client.DescribeRecordFilterList(req)
	if respErr != nil {
		if this.isNotFoundErr(respErr) {
			return nil, nil
		}
		return nil, respErr
	}
	var countRecords = len(resp.Response.RecordList)
	if countRecords == 0 {
		return nil, nil
	}
	for _, recordObj := range resp.Response.RecordList {
		if *recordObj.Name == name && *recordObj.Type == recordType {
			return &dnstypes.Record{
				Id:    types.String(*recordObj.RecordId),
				Name:  *recordObj.Name,
				Type:  *recordObj.Type,
				Value: this.fixCNAME(*recordObj.Type, *recordObj.Value),
				Route: *recordObj.LineId,
				TTL:   types.Int32(*recordObj.TTL),
			}, nil
		}
	}

	return nil, nil
}

// QueryRecords 查询多个记录
func (this *TencentDNSProvider) QueryRecords(domain string, name string, recordType dnstypes.RecordType) ([]*dnstypes.Record, error) {
	// 从缓存中读取
	if this.ProviderId > 0 {
		records, hasRecords, _ := sharedDomainRecordsCache.QueryDomainRecords(this.ProviderId, domain, name, recordType)
		if hasRecords { // 有效的搜索
			return records, nil
		}
	}

	var offset uint64 = 0
	var limit uint64 = 1000
	var records []*dnstypes.Record
	for {
		var req = dnspod.NewDescribeRecordFilterListRequest()
		req.Domain = this.stringVal(domain)
		req.Offset = this.uint64Val(offset)
		req.Limit = this.uint64Val(limit)
		req.SubDomain = this.stringVal(name)
		req.RecordType = []*string{this.stringVal(recordType)}
		resp, respErr := this.client.DescribeRecordFilterList(req)
		if respErr != nil {
			if this.isNotFoundErr(respErr) {
				break
			}
			return nil, respErr
		}
		var countRecords = len(resp.Response.RecordList)
		if countRecords == 0 {
			break
		}
		for _, recordObj := range resp.Response.RecordList {
			records = append(records, &dnstypes.Record{
				Id:    types.String(*recordObj.RecordId),
				Name:  *recordObj.Name,
				Type:  *recordObj.Type,
				Value: this.fixCNAME(*recordObj.Type, *recordObj.Value),
				Route: *recordObj.LineId,
				TTL:   types.Int32(*recordObj.TTL),
			})
		}
		offset += uint64(countRecords)
	}

	return records, nil
}

// AddRecord 设置记录
func (this *TencentDNSProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	if newRecord == nil {
		return errors.New("invalid new record")
	}

	// 在CHANGE记录后面加入点
	if newRecord.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(newRecord.Value, ".") {
		newRecord.Value += "."
	}

	var ttl = newRecord.TTL
	if ttl <= 0 {
		ttl = 600
	}

	var req = dnspod.NewCreateRecordRequest()
	req.Domain = this.stringVal(domain)
	req.SubDomain = this.stringVal(newRecord.Name)
	req.RecordType = this.stringVal(newRecord.Type)
	req.TTL = this.uint64Val(uint64(ttl))
	req.RecordLine = this.stringVal(this.DefaultRouteName()) // 默认必填项，但以RecordLineId优先
	req.RecordLineId = this.stringVal(newRecord.Route)
	req.Value = this.stringVal(newRecord.Value)
	resp, respErr := this.client.CreateRecord(req)
	if respErr != nil {
		return respErr
	}
	newRecord.Id = types.String(*resp.Response.RecordId)

	// 加入缓存
	if this.ProviderId > 0 {
		sharedDomainRecordsCache.AddDomainRecord(this.ProviderId, domain, newRecord)
	}

	return nil
}

// UpdateRecord 修改记录
func (this *TencentDNSProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	if record == nil {
		return errors.New("invalid record")
	}
	if newRecord == nil {
		return errors.New("invalid new record")
	}

	// 在CHANGE记录后面加入点
	if newRecord.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(newRecord.Value, ".") {
		newRecord.Value += "."
	}

	var newRoute = newRecord.Route
	if len(newRoute) == 0 {
		newRoute = this.DefaultRoute()
	}

	var ttl = newRecord.TTL
	if ttl <= 0 {
		ttl = 600
	}

	var req = dnspod.NewModifyRecordRequest()
	req.Domain = this.stringVal(domain)
	req.RecordId = this.uint64Val(types.Uint64(record.Id))
	req.SubDomain = this.stringVal(newRecord.Name)
	req.RecordType = this.stringVal(newRecord.Type)
	req.TTL = this.uint64Val(uint64(ttl))
	req.RecordLine = this.stringVal(this.DefaultRouteName()) // 默认必填项，但以RecordLineId优先
	req.RecordLineId = this.stringVal(newRecord.Route)
	req.Value = this.stringVal(newRecord.Value)
	_, respErr := this.client.ModifyRecord(req)
	if respErr != nil {
		return respErr
	}

	// 修改缓存
	if this.ProviderId > 0 {
		sharedDomainRecordsCache.UpdateDomainRecord(this.ProviderId, domain, newRecord)
	}

	return nil
}

// DeleteRecord 删除记录
func (this *TencentDNSProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	if record == nil {
		return errors.New("invalid record to delete")
	}

	var req = dnspod.NewDeleteRecordRequest()
	req.Domain = this.stringVal(domain)
	req.RecordId = this.uint64Val(types.Uint64(record.Id))
	_, respErr := this.client.DeleteRecord(req)
	if respErr != nil {
		if len(record.Id) > 0 && this.isRecordInvalidErr(respErr) {
			return nil
		}
		return respErr
	}

	// 删除缓存
	if this.ProviderId > 0 {
		sharedDomainRecordsCache.DeleteDomainRecord(this.ProviderId, domain, record.Id)
	}

	return nil
}

// DefaultRoute 默认线路
func (this *TencentDNSProvider) DefaultRoute() string {
	return "0"
}

func (this *TencentDNSProvider) DefaultRouteName() string {
	return "默认"
}

func (this *TencentDNSProvider) fixCNAME(recordType string, recordValue string) string {
	// 修正Record
	if strings.ToUpper(recordType) == dnstypes.RecordTypeCNAME && !strings.HasSuffix(recordValue, ".") {
		recordValue += "."
	}
	return recordValue
}

func (this *TencentDNSProvider) int64Val(v int64) *int64 {
	return &v
}

func (this *TencentDNSProvider) uint64Val(v uint64) *uint64 {
	return &v
}

func (this *TencentDNSProvider) stringVal(s string) *string {
	return &s
}

func (this *TencentDNSProvider) isNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	var sdkErr *tencenterrors.TencentCloudSDKError
	return errors.As(err, &sdkErr) && strings.HasPrefix(sdkErr.Code, "ResourceNotFound.")
}

func (this *TencentDNSProvider) isRecordInvalidErr(err error) bool {
	if err == nil {
		return false
	}
	var sdkErr *tencenterrors.TencentCloudSDKError
	return errors.As(err, &sdkErr) && sdkErr.Code == "InvalidParameter.RecordIdInvalid"
}
