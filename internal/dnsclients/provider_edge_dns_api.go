// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/edgeapi"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var edgeDNSHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

type EdgeDNSAPIProvider struct {
	BaseProvider

	ProviderId int64

	host            string
	accessKeyId     string
	accessKeySecret string

	role                 string // admin | user
	accessToken          string
	accessTokenExpiresAt int64
}

// Auth 认证
func (this *EdgeDNSAPIProvider) Auth(params maps.Map) error {
	this.role = params.GetString("role")
	this.host = params.GetString("host")
	this.accessKeyId = params.GetString("accessKeyId")
	this.accessKeySecret = params.GetString("accessKeySecret")

	if len(this.role) == 0 {
		this.role = "user"
	}

	if len(this.host) == 0 {
		return errors.New("'host' should not be empty")
	}
	if !regexp.MustCompile(`^(?i)(http|https):`).MatchString(this.host) {
		this.host = "http://" + this.host
	}

	if len(this.accessKeyId) == 0 {
		return errors.New("'accessKeyId' should not be empty")
	}
	if len(this.accessKeySecret) == 0 {
		return errors.New("'accessKeySecret' should not be empty")
	}

	return nil
}

// GetDomains 获取所有域名列表
func (this *EdgeDNSAPIProvider) GetDomains() (domains []string, err error) {
	var offset = 0
	var size = 100
	for {
		var resp = &edgeapi.ListNSDomainsResponse{}
		err = this.doAPI("/NSDomainService/ListNSDomains", map[string]any{
			"offset": offset,
			"size":   size,
		}, resp)
		if err != nil {
			return
		}

		for _, domain := range resp.Data.NSDomains {
			domains = append(domains, domain.Name)
		}

		if len(resp.Data.NSDomains) < size {
			break
		}

		offset += size
	}

	return
}

// GetRecords 获取域名解析记录列表
func (this *EdgeDNSAPIProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	var domainResp = &edgeapi.FindDomainWithNameResponse{}
	err = this.doAPI("/NSDomainService/FindNSDomainWithName", map[string]any{
		"name": domain,
	}, domainResp)
	if err != nil {
		return nil, err
	}

	var domainId = domainResp.Data.NSDomain.Id
	if domainId == 0 {
		return nil, nil
	}

	var offset = 0
	var size = 100
	for {
		var recordsResp = &edgeapi.ListNSRecordsResponse{}
		err = this.doAPI("/NSRecordService/ListNSRecords", map[string]any{
			"nsDomainId": domainId,
			"offset":     offset,
			"size":       size,
		}, recordsResp)
		if err != nil {
			return nil, err
		}

		var nsRecords = recordsResp.Data.NSRecords
		for _, record := range nsRecords {
			var routeCode = this.DefaultRoute()
			if len(record.NSRoutes) > 0 {
				routeCode = record.NSRoutes[0].Code
			}

			records = append(records, &dnstypes.Record{
				Id:    types.String(record.Id),
				Name:  record.Name,
				Type:  record.Type,
				Value: record.Value,
				Route: routeCode,
				TTL:   record.TTL,
			})
		}

		if len(nsRecords) < size {
			break
		}

		offset += size
	}

	return
}

// GetRoutes 读取域名支持的线路数据
func (this *EdgeDNSAPIProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	// default
	routes = append(routes, &dnstypes.Route{
		Name: "默认线路",
		Code: this.DefaultRoute(),
	})

	// 世界区域
	{
		var routesResp = &edgeapi.FindAllNSRoutesResponse{}
		err = this.doAPI("/NSRouteService/FindAllDefaultWorldRegionRoutes", map[string]any{}, routesResp)
		if err != nil {
			return nil, err
		}
		for _, route := range routesResp.Data.NSRoutes {
			routes = append(routes, &dnstypes.Route{
				Name: route.Name,
				Code: route.Code,
			})
		}
	}

	// 中国省份
	{
		var routesResp = &edgeapi.FindAllNSRoutesResponse{}
		err = this.doAPI("/NSRouteService/FindAllDefaultChinaProvinceRoutes", map[string]any{}, routesResp)
		if err != nil {
			return nil, err
		}
		for _, route := range routesResp.Data.NSRoutes {
			routes = append(routes, &dnstypes.Route{
				Name: route.Name,
				Code: route.Code,
			})
		}
	}

	// ISP
	{
		var routesResp = &edgeapi.FindAllNSRoutesResponse{}
		err = this.doAPI("/NSRouteService/FindAllDefaultISPRoutes", map[string]any{}, routesResp)
		if err != nil {
			return nil, err
		}
		for _, route := range routesResp.Data.NSRoutes {
			routes = append(routes, &dnstypes.Route{
				Name: route.Name,
				Code: route.Code,
			})
		}
	}

	// 自定义
	{
		var routesResp = &edgeapi.FindAllNSRoutesResponse{}
		err = this.doAPI("/NSRouteService/FindAllNSRoutes", map[string]any{}, routesResp)
		if err != nil {
			return nil, err
		}
		for _, route := range routesResp.Data.NSRoutes {
			routes = append(routes, &dnstypes.Route{
				Name: route.Name,
				Code: route.Code,
			})
		}
	}

	return
}

// QueryRecord 查询单个记录
func (this *EdgeDNSAPIProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	var domainResp = &edgeapi.FindDomainWithNameResponse{}
	err := this.doAPI("/NSDomainService/FindNSDomainWithName", map[string]any{
		"name": domain,
	}, domainResp)
	if err != nil {
		return nil, err
	}

	var domainId = domainResp.Data.NSDomain.Id
	if domainId == 0 {
		return nil, errors.New("can not find domain '" + domain + "'")
	}

	var recordResp = &edgeapi.FindNSRecordWithNameAndTypeResponse{}
	err = this.doAPI("/NSRecordService/FindNSRecordWithNameAndType", map[string]any{
		"nsDomainId": domainId,
		"name":       name,
		"type":       recordType,
	}, recordResp)
	if err != nil {
		return nil, err
	}

	var record = recordResp.Data.NSRecord
	if record.Id <= 0 {
		return nil, nil
	}

	var routeCode = this.DefaultRoute()
	if len(record.NSRoutes) > 0 {
		routeCode = record.NSRoutes[0].Code
	}

	return &dnstypes.Record{
		Id:    types.String(record.Id),
		Name:  record.Name,
		Type:  record.Type,
		Value: record.Value,
		Route: routeCode,
		TTL:   record.TTL,
	}, nil
}

// AddRecord 设置记录
func (this *EdgeDNSAPIProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	var domainResp = &edgeapi.FindDomainWithNameResponse{}
	err := this.doAPI("/NSDomainService/FindNSDomainWithName", map[string]any{
		"name": domain,
	}, domainResp)
	if err != nil {
		return err
	}

	var domainId = domainResp.Data.NSDomain.Id
	if domainId == 0 {
		return errors.New("can not find domain '" + domain + "'")
	}

	if newRecord.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(newRecord.Value, ".") {
		newRecord.Value += "."
	}

	var createResp = &edgeapi.CreateNSRecordResponse{}
	var routes = []string{}
	if len(newRecord.Route) > 0 {
		routes = []string{newRecord.Route}
	}
	err = this.doAPI("/NSRecordService/CreateNSRecord", map[string]any{
		"nsDomainId":   domainId,
		"name":         newRecord.Name,
		"type":         strings.ToUpper(newRecord.Type),
		"value":        newRecord.Value,
		"ttl":          newRecord.TTL,
		"nsRouteCodes": routes,
	}, createResp)

	if err != nil {
		return err
	}

	newRecord.Id = types.String(createResp.Data.NSRecordId)

	return nil
}

// UpdateRecord 修改记录
func (this *EdgeDNSAPIProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	if newRecord.Type == dnstypes.RecordTypeCNAME && !strings.HasSuffix(newRecord.Value, ".") {
		newRecord.Value += "."
	}

	var createResp = &edgeapi.UpdateNSRecordResponse{}
	var routes = []string{}
	if len(newRecord.Route) > 0 {
		routes = []string{newRecord.Route}
	}
	err := this.doAPI("/NSRecordService/UpdateNSRecord", map[string]any{
		"nsRecordId":   types.Int64(record.Id),
		"name":         newRecord.Name,
		"type":         strings.ToUpper(newRecord.Type),
		"value":        newRecord.Value,
		"ttl":          newRecord.TTL,
		"nsRouteCodes": routes,
		"isOn":         true, // important
	}, createResp)

	return err
}

// DeleteRecord 删除记录
func (this *EdgeDNSAPIProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	var resp = &edgeapi.SuccessResponse{}
	err := this.doAPI("/NSRecordService/DeleteNSRecord", map[string]any{
		"nsRecordId": types.Int64(record.Id),
	}, resp)
	return err
}

// DefaultRoute 默认线路
func (this *EdgeDNSAPIProvider) DefaultRoute() string {
	return "default"
}

func (this *EdgeDNSAPIProvider) doAPI(path string, params map[string]any, respPtr edgeapi.ResponseInterface) error {
	accessToken, err := this.getToken()
	if err != nil {
		return err
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, this.host+path, bytes.NewReader(paramsJSON))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)
	req.Header.Set("X-Edge-Access-Token", accessToken)

	resp, err := edgeDNSHTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("invalid response status code '" + types.String(resp.StatusCode) + "'")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, respPtr)
	if err != nil {
		return errors.New("decode response failed: " + err.Error() + ", JSON: " + string(data))
	}

	if !respPtr.IsValid() {
		return respPtr.Error()
	}

	return err
}

func (this *EdgeDNSAPIProvider) getToken() (string, error) {
	if len(this.accessToken) > 0 && this.accessTokenExpiresAt > time.Now().Unix()+600 /** 600秒是防止当前服务器和API服务器之间有时间差 **/ {
		return this.accessToken, nil
	}

	var params = maps.Map{
		"type":        this.role,
		"accessKeyId": this.accessKeyId,
		"accessKey":   this.accessKeySecret,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, this.host+"/APIAccessTokenService/getAPIAccessToken", bytes.NewReader(paramsJSON))
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", teaconst.ProductName+"/"+teaconst.Version)
	resp, err := edgeDNSHTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("invalid response code '" + types.String(resp.StatusCode) + "'")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp = &edgeapi.GetAPIAccessToken{}
	err = json.Unmarshal(data, tokenResp)
	if err != nil {
		return "", err
	}

	if tokenResp.Code != 200 {
		return "", errors.New("invalid code '" + types.String(tokenResp.Code) + "', message: " + tokenResp.Message)
	}

	this.accessToken = tokenResp.Data.Token
	this.accessTokenExpiresAt = tokenResp.Data.ExpiresAt
	return this.accessToken, nil
}
