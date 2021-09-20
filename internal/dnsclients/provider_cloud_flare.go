// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/cloudflare"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const CloudFlareAPIEndpoint = "https://api.cloudflare.com/client/v4/"
const CloudFlareDefaultRoute = "default"

var cloudFlareHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

type CloudFlareProvider struct {
	BaseProvider

	apiKey string // API密钥
	email  string // 账号邮箱

	zoneMap    map[string]string // domain => zoneId
	zoneLocker sync.Mutex
}

// Auth 认证
func (this *CloudFlareProvider) Auth(params maps.Map) error {
	this.apiKey = params.GetString("apiKey")
	if len(this.apiKey) == 0 {
		return errors.New("'apiKey' should not be empty")
	}

	this.email = params.GetString("email")
	if len(this.email) == 0 {
		return errors.New("'email' should not be empty")
	}

	this.zoneMap = map[string]string{}

	return nil
}

// GetDomains 获取所有域名列表
func (this *CloudFlareProvider) GetDomains() (domains []string, err error) {
	resp := new(cloudflare.ZonesResponse)
	err = this.doAPI(http.MethodGet, "zones", map[string]string{}, nil, resp)
	if err != nil {
		return nil, err
	}

	for _, zone := range resp.Result {
		domains = append(domains, zone.Name)
	}

	return
}

// GetRecords 获取域名解析记录列表
func (this *CloudFlareProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return nil, err
	}

	// 这个页数限制预示着每次最多只能获取 500 * 100 即5万个数据
	for page := 1; page <= 500; page++ {
		resp := new(cloudflare.GetDNSRecordsResponse)
		err = this.doAPI(http.MethodGet, "zones/"+zoneId+"/dns_records", map[string]string{
			"per_page": "100",
			"page":     strconv.Itoa(page),
		}, nil, resp)
		if err != nil {
			return nil, err
		}
		if len(resp.Result) == 0 {
			break
		}

		for _, record := range resp.Result {
			// 修正Record
			if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Content, ".") {
				record.Content += "."
			}

			record.Name = strings.TrimSuffix(record.Name, "."+domain)

			records = append(records, &dnstypes.Record{
				Id:    record.Id,
				Name:  record.Name,
				Type:  record.Type,
				Value: record.Content,
				Route: CloudFlareDefaultRoute,
			})
		}
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *CloudFlareProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	routes = []*dnstypes.Route{
		{Name: "默认", Code: CloudFlareDefaultRoute},
	}
	return
}

// QueryRecord 查询单个记录
func (this *CloudFlareProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return nil, err
	}

	resp := new(cloudflare.GetDNSRecordsResponse)
	err = this.doAPI(http.MethodGet, "zones/"+zoneId+"/dns_records", map[string]string{
		"per_page": "100",
		"name":     name + "." + domain,
		"type":     recordType,
	}, nil, resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Result) == 0 {
		return nil, nil
	}

	record := resp.Result[0]

	// 修正Record
	if record.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(record.Content, ".") {
		record.Content += "."
	}

	record.Name = strings.TrimSuffix(record.Name, "."+domain)

	return &dnstypes.Record{
		Id:    record.Id,
		Name:  record.Name,
		Type:  record.Type,
		Value: record.Content,
		TTL:   types.Int32(record.Ttl),
		Route: CloudFlareDefaultRoute,
	}, nil
}

// AddRecord 设置记录
func (this *CloudFlareProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return err
	}

	resp := new(cloudflare.CreateDNSRecordResponse)

	var ttl = newRecord.TTL
	if ttl <= 0 {
		ttl = 1 // 自动默认
	}

	err = this.doAPI(http.MethodPost, "zones/"+zoneId+"/dns_records", nil, maps.Map{
		"type":    newRecord.Type,
		"name":    newRecord.Name + "." + domain,
		"content": newRecord.Value,
		"ttl":     ttl,
	}, resp)
	if err != nil {
		return err
	}
	return nil
}

// UpdateRecord 修改记录
func (this *CloudFlareProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return err
	}

	var ttl = newRecord.TTL
	if ttl <= 0 {
		ttl = 1 // 自动默认
	}

	resp := new(cloudflare.UpdateDNSRecordResponse)
	return this.doAPI(http.MethodPut, "zones/"+zoneId+"/dns_records/"+record.Id, nil, maps.Map{
		"type":    newRecord.Type,
		"name":    newRecord.Name + "." + domain,
		"content": newRecord.Value,
		"ttl":     ttl,
	}, resp)
}

// DeleteRecord 删除记录
func (this *CloudFlareProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	zoneId, err := this.findZoneIdWithDomain(domain)
	if err != nil {
		return err
	}

	resp := new(cloudflare.DeleteDNSRecordResponse)
	err = this.doAPI(http.MethodDelete, "zones/"+zoneId+"/dns_records/"+record.Id, map[string]string{}, nil, resp)
	if err != nil {
		return err
	}
	return nil
}

// DefaultRoute 默认线路
func (this *CloudFlareProvider) DefaultRoute() string {
	return CloudFlareDefaultRoute
}

// 执行API
func (this *CloudFlareProvider) doAPI(method string, apiPath string, args map[string]string, bodyMap maps.Map, respPtr cloudflare.ResponseInterface) error {
	apiURL := CloudFlareAPIEndpoint + strings.TrimLeft(apiPath, "/")
	if len(args) > 0 {
		apiURL += "?"
		argStrings := []string{}
		for k, v := range args {
			argStrings = append(argStrings, k+"="+url.QueryEscape(v))
		}
		apiURL += strings.Join(argStrings, "&")
	}
	method = strings.ToUpper(method)

	var bodyReader io.Reader = nil
	if bodyMap != nil {
		bodyData, err := json.Marshal(bodyMap)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	req, err := http.NewRequest(method, apiURL, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Key", this.apiKey)
	req.Header.Set("x-Auth-Email", this.email)
	resp, err := cloudFlareHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 0 {
		return errors.New("invalid response status '" + strconv.Itoa(resp.StatusCode) + "', response '" + string(data) + "'")
	}

	err = json.Unmarshal(data, respPtr)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("response error: " + string(data))
	}

	return nil
}

// 列出一个域名对应的区域
func (this *CloudFlareProvider) findZoneIdWithDomain(domain string) (zoneId string, err error) {
	this.zoneLocker.Lock()
	cacheZonedId, ok := this.zoneMap[domain]
	if ok {
		this.zoneLocker.Unlock()
		return cacheZonedId, nil
	}
	this.zoneLocker.Unlock()

	resp := new(cloudflare.ZonesResponse)
	err = this.doAPI(http.MethodGet, "zones", map[string]string{
		"name": domain,
	}, nil, resp)
	if err != nil {
		return "", err
	}
	if len(resp.Result) == 0 {
		return "", errors.New("can not found zone for domain '" + domain + "'")
	}
	zoneId = resp.Result[0].Id
	this.zoneLocker.Lock()
	this.zoneMap[domain] = zoneId
	this.zoneLocker.Unlock()
	return zoneId, nil
}
