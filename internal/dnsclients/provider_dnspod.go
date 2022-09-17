package dnsclients

import (
	"crypto/tls"
	"encoding/json"
	"errors"
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
type DNSPodProvider struct {
	BaseProvider

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
	var size = 100

	for {
		domainsResp, err := this.post("/Domain.List", map[string]string{
			"offset": numberutils.FormatInt(offset),
			"length": numberutils.FormatInt(size),
		})
		if err != nil {
			return nil, err
		}
		offset += size

		var domainsSlice = domainsResp.GetSlice("domains")
		if len(domainsSlice) == 0 {
			break
		}

		for _, domain := range domainsSlice {
			domainMap := maps.NewMap(domain)
			domains = append(domains, domainMap.GetString("name"))
		}

		// 检查是否到头
		var info = domainsResp.GetMap("info")
		var recordTotal = info.GetInt("all_total")
		if offset >= recordTotal {
			break
		}
	}
	return
}

// GetRecords 获取域名列表
func (this *DNSPodProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	var offset = 0
	var size = 100
	for {
		recordsResp, err := this.post("/Record.List", map[string]string{
			"domain": domain,
			"offset": numberutils.FormatInt(offset),
			"length": numberutils.FormatInt(size),
		})
		if err != nil {
			return nil, err
		}
		offset += size

		// 记录
		var recordSlice = recordsResp.GetSlice("records")
		for _, record := range recordSlice {
			recordMap := maps.NewMap(record)
			records = append(records, &dnstypes.Record{
				Id:    recordMap.GetString("id"),
				Name:  recordMap.GetString("name"),
				Type:  recordMap.GetString("type"),
				Value: recordMap.GetString("value"),
				Route: recordMap.GetString("line"),
				TTL:   recordMap.GetInt32("ttl"),
			})
		}

		// 检查是否到头
		var info = recordsResp.GetMap("info")
		var recordTotal = info.GetInt("record_total")
		if offset >= recordTotal {
			break
		}
	}
	return
}

// GetRoutes 读取线路数据
func (this *DNSPodProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	infoResp, err := this.post("/Domain.Info", map[string]string{
		"domain": domain,
	})
	if err != nil {
		return nil, err
	}
	domainInfo := infoResp.GetMap("domain")
	grade := domainInfo.GetString("grade")

	linesResp, err := this.post("/Record.Line", map[string]string{
		"domain":       domain,
		"domain_grade": grade,
	})
	if err != nil {
		return nil, err
	}

	lines := linesResp.GetSlice("lines")
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

	return routes, nil
}

// QueryRecord 查询单个记录
func (this *DNSPodProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
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
	_, err := this.post("/Record.Create", args)
	return this.WrapError(err, domain, newRecord)
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
	_, err := this.post("/Record.Modify", args)
	return this.WrapError(err, domain, newRecord)
}

// DeleteRecord 删除记录
func (this *DNSPodProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	if record == nil {
		return errors.New("invalid record to delete")
	}

	_, err := this.post("/Record.Remove", map[string]string{
		"domain":    domain,
		"record_id": record.Id,
	})

	return this.WrapError(err, domain, record)
}

// 发送请求
func (this *DNSPodProvider) post(path string, params map[string]string) (maps.Map, error) {
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
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GoEdge-Client/1.0.0 (iwind.liu@gmail.com)")
	req.Header.Set("Accept", "*/*")

	resp, err := dnsPodHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var m = maps.Map{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}
	var status = m.GetMap("status")
	var code = status.GetString("code")
	if code != "1" {
		return nil, errors.New("API response error: code: " + code + ", message: " + status.GetString("message"))
	}

	return m, nil
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
