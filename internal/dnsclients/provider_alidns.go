package dnsclients

import (
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/iwind/TeaGo/maps"
	"strings"
)

// 阿里云服务商
type AliDNSProvider struct {
	accessKeyId     string
	accessKeySecret string
}

// 认证
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

// 获取域名列表
func (this *AliDNSProvider) GetRecords(domain string) (records []*Record, err error) {
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
			if record.Type == RecordTypeCName && !strings.HasSuffix(record.Value, ".") {
				record.Value += "."
			}

			records = append(records, &Record{
				Id:    record.RecordId,
				Name:  record.RR,
				Type:  record.Type,
				Value: record.Value,
				Route: record.Line,
			})
		}

		pageNumber++
		if int64((pageNumber-1)*size) >= resp.TotalCount {
			break
		}
	}

	return
}

// 读取域名支持的线路数据
func (this *AliDNSProvider) GetRoutes(domain string) (routes []*Route, err error) {
	req := alidns.CreateDescribeSupportLinesRequest()
	req.DomainName = domain

	resp := alidns.CreateDescribeSupportLinesResponse()
	err = this.doAPI(req, resp)
	if err != nil {
		return nil, err
	}
	for _, line := range resp.RecordLines.RecordLine {
		routes = append(routes, &Route{
			Name: line.LineName,
			Code: line.LineCode,
		})
	}
	return
}

// 设置记录
func (this *AliDNSProvider) AddRecord(domain string, newRecord *Record) error {
	req := alidns.CreateAddDomainRecordRequest()
	req.RR = newRecord.Name
	req.Type = newRecord.Type
	req.Value = newRecord.Value
	req.DomainName = domain
	req.Line = newRecord.Route

	resp := alidns.CreateAddDomainRecordResponse()
	err := this.doAPI(req, resp)
	if err != nil {
		return err
	}
	if resp.IsSuccess() {
		return nil
	}

	return errors.New(resp.GetHttpContentString())
}

// 修改记录
func (this *AliDNSProvider) UpdateRecord(domain string, record *Record, newRecord *Record) error {
	req := alidns.CreateUpdateDomainRecordRequest()
	req.RecordId = record.Id
	req.RR = newRecord.Name
	req.Type = newRecord.Type
	req.Value = newRecord.Value
	req.Line = newRecord.Route

	resp := alidns.CreateUpdateDomainRecordResponse()
	err := this.doAPI(req, resp)
	return err
}

// 删除记录
func (this *AliDNSProvider) DeleteRecord(domain string, record *Record) error {
	req := alidns.CreateDeleteDomainRecordRequest()
	req.RecordId = record.Id

	resp := alidns.CreateDeleteDomainRecordResponse()
	err := this.doAPI(req, resp)
	return err
}

// 默认线路
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
