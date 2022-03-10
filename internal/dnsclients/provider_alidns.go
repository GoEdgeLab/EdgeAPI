package dnsclients

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"strings"
)

// AliDNSProvider 阿里云服务商
type AliDNSProvider struct {
	BaseProvider

	accessKeyId     string
	accessKeySecret string
}

// Auth 认证
func (this *AliDNSProvider) Auth(params maps.Map) error {
	this.accessKeyId = params.GetString("accessKeyId")
	this.accessKeySecret = params.GetString("accessKeySecret")
	if len(this.accessKeyId) == 0 {
		return errors.New("'accessKeyId' should not be empty")
	}
	if len(this.accessKeySecret) == 0 {
		return errors.New("'accessKeySecret' should not be empty")
	}
	return nil
}

// GetDomains 获取所有域名列表
func (this *AliDNSProvider) GetDomains() (domains []string, err error) {
	pageNumber := 1
	size := 100

	for {
		req := alidns.CreateDescribeDomainsRequest()
		req.PageNumber = requests.NewInteger(pageNumber)
		req.PageSize = requests.NewInteger(size)
		resp := alidns.CreateDescribeDomainsResponse()
		err = this.doAPI(req, resp)
		if err != nil {
			return nil, err
		}

		for _, domain := range resp.Domains.Domain {
			domains = append(domains, domain.DomainName)
		}

		pageNumber++
		if int64((pageNumber-1)*size) >= resp.TotalCount {
			break
		}
	}

	return
}

// GetRecords 获取域名列表
func (this *AliDNSProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	pageNumber := 1
	size := 100

	for {
		req := alidns.CreateDescribeDomainRecordsRequest()
		req.DomainName = domain
		req.PageNumber = requests.NewInteger(pageNumber)
		req.PageSize = requests.NewInteger(size)

		resp := alidns.CreateDescribeDomainRecordsResponse()
		err = this.doAPI(req, resp)
		if err != nil {
			return nil, err
		}
		for _, record := range resp.DomainRecords.Record {
			// 修正Record
			if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Value, ".") {
				record.Value += "."
			}

			records = append(records, &dnstypes.Record{
				Id:    record.RecordId,
				Name:  record.RR,
				Type:  record.Type,
				Value: record.Value,
				Route: record.Line,
				TTL:   types.Int32(record.TTL),
			})
		}

		pageNumber++
		if int64((pageNumber-1)*size) >= resp.TotalCount {
			break
		}
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *AliDNSProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	req := alidns.CreateDescribeSupportLinesRequest()
	req.DomainName = domain

	resp := alidns.CreateDescribeSupportLinesResponse()
	err = this.doAPI(req, resp)
	if err != nil {
		return nil, err
	}
	for _, line := range resp.RecordLines.RecordLine {
		routes = append(routes, &dnstypes.Route{
			Name: line.LineName,
			Code: line.LineCode,
		})
	}
	return
}

// QueryRecord 查询单个记录
func (this *AliDNSProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	records, err := this.GetRecords(domain)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.Name == name && record.Type == recordType {
			return record, nil
		}
	}
	return nil, err
}

// AddRecord 设置记录
func (this *AliDNSProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	req := alidns.CreateAddDomainRecordRequest()
	req.RR = newRecord.Name
	req.Type = newRecord.Type
	req.Value = newRecord.Value
	req.DomainName = domain
	req.Line = newRecord.Route

	if newRecord.TTL > 0 {
		req.TTL = requests.NewInteger(types.Int(newRecord.TTL))
	}

	resp := alidns.CreateAddDomainRecordResponse()
	err := this.doAPI(req, resp)
	if err != nil {
		return this.WrapError(err, domain, newRecord)
	}
	if resp.IsSuccess() {
		return nil
	}

	return this.WrapError(errors.New(resp.GetHttpContentString()), domain, newRecord)
}

// UpdateRecord 修改记录
func (this *AliDNSProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	req := alidns.CreateUpdateDomainRecordRequest()
	req.RecordId = record.Id
	req.RR = newRecord.Name
	req.Type = newRecord.Type
	req.Value = newRecord.Value
	req.Line = newRecord.Route

	if newRecord.TTL > 0 {
		req.TTL = requests.NewInteger(types.Int(newRecord.TTL))
	}

	resp := alidns.CreateUpdateDomainRecordResponse()
	err := this.doAPI(req, resp)
	return this.WrapError(err, domain, newRecord)
}

// DeleteRecord 删除记录
func (this *AliDNSProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	req := alidns.CreateDeleteDomainRecordRequest()
	req.RecordId = record.Id

	resp := alidns.CreateDeleteDomainRecordResponse()
	err := this.doAPI(req, resp)
	return this.WrapError(err, domain, record)
}

// DefaultRoute 默认线路
func (this *AliDNSProvider) DefaultRoute() string {
	return "default"
}

// 执行请求
func (this *AliDNSProvider) doAPI(req requests.AcsRequest, resp responses.AcsResponse) error {
	req.SetScheme("https")

	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", this.accessKeyId, this.accessKeySecret)
	if err != nil {
		return err
	}
	err = client.DoAction(req, resp)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.GetHttpContentString())
	}
	return nil
}
