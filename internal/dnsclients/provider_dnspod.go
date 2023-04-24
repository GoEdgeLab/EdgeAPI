package dnsclients

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnspod"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DNSPodMaxTTL        int32 = 604800
	DNSPodInternational       = "international"
)

var dnsPodHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// DNSPodProvider DNSPod服务商
// TODO 考虑支持线路ID
// TODO 支持自定义线路
type DNSPodProvider struct {
	BaseProvider

	ProviderId int64

	region   string
	apiId    string
	apiToken string
}

// Auth 认证
func (this *DNSPodProvider) Auth(params maps.Map) error {
	this.apiId = params.GetString("id")
	this.apiToken = params.GetString("token")
	this.region = params.GetString("region")

	if len(this.apiId) == 0 {
		return errors.New("'id' should be not empty")
	}
	if len(this.apiToken) == 0 {
		return errors.New("'token' should not be empty")
	}

	return nil
}

// GetDomains 获取所有域名列表
func (this *DNSPodProvider) GetDomains() (domains []string, err error) {
	var offset = 0
	var size = 3000

	for {
		var resp = new(dnspod.DomainListResponse)

		err := this.doAPI("/Domain.List", map[string]string{
			"offset": numberutils.FormatInt(offset),
			"length": numberutils.FormatInt(size),
		}, resp)
		if err != nil {
			return nil, err
		}
		offset += size

		for _, domain := range resp.Domains {
			domains = append(domains, domain.Name)
		}

		// 检查是否到头
		var recordTotal = resp.Info.AllTotal
		if offset >= recordTotal {
			break
		}
	}
	return
}

// GetRecords 获取域名列表
func (this *DNSPodProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	var offset = 0
	var size = 3000
	for {
		var resp = new(dnspod.RecordListResponse)
		err := this.doAPI("/Record.List", map[string]string{
			"domain": domain,
			"offset": numberutils.FormatInt(offset),
			"length": numberutils.FormatInt(size),
		}, resp)
		if err != nil {
			return nil, err
		}
		offset += size

		// 记录
		for _, record := range resp.Records {
			records = append(records, &dnstypes.Record{
				Id:    types.String(record.Id),
				Name:  record.Name,
				Type:  record.Type,
				Value: record.Value,
				Route: record.Line,
				TTL:   types.Int32(record.TTL),
			})
		}

		// 检查是否到头
		var recordTotal = types.Int(resp.Info.RecordTotal)
		if offset >= recordTotal {
			break
		}
	}

	// 写入缓存
	if this.ProviderId > 0 {
		sharedDomainRecordsCache.WriteDomainRecords(this.ProviderId, domain, records)
	}

	return
}

// GetRoutes 读取线路数据
func (this *DNSPodProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	var domainInfoResp = new(dnspod.DomainInfoResponse)
	err = this.doAPI("/Domain.Info", map[string]string{
		"domain": domain,
	}, domainInfoResp)
	if err != nil {
		return nil, err
	}
	var grade = domainInfoResp.Domain.Grade

	var linesResp = new(dnspod.RecordLineResponse)
	err = this.doAPI("/Record.Line", map[string]string{
		"domain":       domain,
		"domain_grade": grade,
	}, linesResp)
	if err != nil {
		return nil, err
	}

	var lines = linesResp.Lines
	if len(lines) == 0 {
		return nil, nil
	}
	for _, line := range lines {
		lineString := types.String(line)
		routes = append(routes, &dnstypes.Route{
			Name: lineString,
			Code: lineString,
		})
	}

	// 自定义线路分组
	var groupsPerPage = 100
	var customLineGroupListResponse = new(dnspod.CustomLineGroupListResponse)
	err = this.doAPI("/Custom.Line.Group.List", map[string]string{
		"domain": domain,
		"offset": "0",
		"length": types.String(groupsPerPage),
	}, customLineGroupListResponse)
	if err != nil {
		// 忽略自定义分组错误
		err = nil
	} else {
		if len(customLineGroupListResponse.Data.LineGroups) > 0 {
			for _, lineGroup := range customLineGroupListResponse.Data.LineGroups {
				routes = append(routes, &dnstypes.Route{
					Name: "分组：" + lineGroup.Name,
					Code: lineGroup.Name,
				})
			}
		}

		if customLineGroupListResponse.Data.Info.Total > customLineGroupListResponse.Data.Info.NowTotal {
			for page := 1; page <= 100; /** 最多100页 **/ page++ {
				err = this.doAPI("/Custom.Line.Group.List", map[string]string{
					"domain": domain,
					"offset": types.String(page * groupsPerPage),
					"length": types.String(groupsPerPage),
				}, customLineGroupListResponse)
				if err != nil {
					// 忽略错误
					err = nil
				} else {
					if len(customLineGroupListResponse.Data.LineGroups) == 0 {
						break
					}

					for _, lineGroup := range customLineGroupListResponse.Data.LineGroups {
						routes = append(routes, &dnstypes.Route{
							Name: "分组：" + lineGroup.Name,
							Code: lineGroup.Name,
						})
					}
				}
			}
		}
	}

	return routes, nil
}

// QueryRecord 查询单个记录
func (this *DNSPodProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	// 从缓存中读取
	if this.ProviderId > 0 {
		record, hasRecords, _ := sharedDomainRecordsCache.QueryDomainRecord(this.ProviderId, domain, name, recordType)
		if hasRecords { // 有效的搜索
			return record, nil
		}
	}

	records, err := this.GetRecords(domain)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.Name == name && record.Type == recordType {
			return record, nil
		}
	}
	return nil, nil
}

// QueryRecords 查询多个记录
func (this *DNSPodProvider) QueryRecords(domain string, name string, recordType dnstypes.RecordType) ([]*dnstypes.Record, error) {
	// 从缓存中读取
	if this.ProviderId > 0 {
		records, hasRecords, _ := sharedDomainRecordsCache.QueryDomainRecords(this.ProviderId, domain, name, recordType)
		if hasRecords { // 有效的搜索
			return records, nil
		}
	}

	records, err := this.GetRecords(domain)
	if err != nil {
		return nil, err
	}
	var result = []*dnstypes.Record{}
	for _, record := range records {
		if record.Name == name && record.Type == recordType {
			result = append(result, record)
		}
	}
	return result, nil
}

// AddRecord 设置记录
func (this *DNSPodProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	if newRecord == nil {
		return errors.New("invalid new record")
	}

	// 在CHANGE记录后面加入点
	if newRecord.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(newRecord.Value, ".") {
		newRecord.Value += "."
	}

	var args = map[string]string{
		"domain":      domain,
		"sub_domain":  newRecord.Name,
		"record_type": newRecord.Type,
		"value":       newRecord.Value,
		"record_line": newRecord.Route,
	}
	if newRecord.TTL > 0 && newRecord.TTL <= DNSPodMaxTTL {
		args["ttl"] = types.String(newRecord.TTL)
	}
	var resp = new(dnspod.RecordCreateResponse)
	err := this.doAPI("/Record.Create", args, resp)
	if err != nil {
		return this.WrapError(err, domain, newRecord)
	}
	newRecord.Id = types.String(resp.Record.Id)

	// 加入缓存
	if this.ProviderId > 0 {
		sharedDomainRecordsCache.AddDomainRecord(this.ProviderId, domain, newRecord)
	}

	return nil
}

// UpdateRecord 修改记录
func (this *DNSPodProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
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

	var args = map[string]string{
		"domain":      domain,
		"record_id":   record.Id,
		"sub_domain":  newRecord.Name,
		"record_type": newRecord.Type,
		"value":       newRecord.Value,
		"record_line": newRecord.Route,
	}
	if newRecord.TTL > 0 && newRecord.TTL <= DNSPodMaxTTL {
		args["ttl"] = types.String(newRecord.TTL)
	}
	var resp = new(dnspod.RecordModifyResponse)
	err := this.doAPI("/Record.Modify", args, resp)
	if err != nil {
		return this.WrapError(err, domain, newRecord)
	}

	newRecord.Id = record.Id

	// 修改缓存
	if this.ProviderId > 0 {
		sharedDomainRecordsCache.UpdateDomainRecord(this.ProviderId, domain, newRecord)
	}

	return nil
}

// DeleteRecord 删除记录
func (this *DNSPodProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	if record == nil {
		return errors.New("invalid record to delete")
	}

	var resp = new(dnspod.RecordRemoveResponse)
	err := this.doAPI("/Record.Remove", map[string]string{
		"domain":    domain,
		"record_id": record.Id,
	}, resp)

	if err != nil {
		return this.WrapError(err, domain, record)
	}

	// 删除缓存
	if this.ProviderId > 0 {
		sharedDomainRecordsCache.DeleteDomainRecord(this.ProviderId, domain, record.Id)
	}

	return nil
}

// 发送请求
func (this *DNSPodProvider) doAPI(path string, params map[string]string, respPtr dnspod.ResponseInterface) error {
	var apiHost = "https://dnsapi.cn"
	var lang = "cn"
	if this.isInternational() { // 国际版
		apiHost = "https://api.dnspod.com"
		lang = "en"
	}
	var query = url.Values{
		"login_token": []string{this.apiId + "," + this.apiToken},
		"format":      []string{"json"},
		"lang":        []string{lang},
	}

	for p, v := range params {
		query[p] = []string{v}
	}

	req, err := http.NewRequest(http.MethodPost, apiHost+path, strings.NewReader(query.Encode()))
	if err != nil {
		return errors.New("create request failed: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GoEdge-Client/1.0.0 (iwind.liu@gmail.com)")
	req.Header.Set("Accept", "*/*")

	resp, err := dnsPodHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &respPtr)
	if err != nil {
		return err
	}
	if !respPtr.IsOk() {
		code, message := respPtr.LastError()
		return errors.New("API response error: code: " + code + ", message: " + message)
	}

	return nil
}

// DefaultRoute 默认线路
func (this *DNSPodProvider) DefaultRoute() string {
	if this.isInternational() {
		return "Default"
	}
	return "默认"
}

func (this *DNSPodProvider) isInternational() bool {
	return this.region == DNSPodInternational
}
