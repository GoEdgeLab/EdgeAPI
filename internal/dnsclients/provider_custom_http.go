package dnsclients

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/json"
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/maps"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var customHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// CustomHTTPProvider HTTP自定义DNS
type CustomHTTPProvider struct {
	url    string
	secret string
}

// Auth 认证
// 参数：
//   - url
//   - secret
func (this *CustomHTTPProvider) Auth(params maps.Map) error {
	this.url = params.GetString("url")
	if len(this.url) == 0 {
		return errors.New("'url' should not be empty")
	}

	this.secret = params.GetString("secret")
	if len(this.secret) == 0 {
		return errors.New("'secret' should not be empty")
	}

	return nil
}

// GetDomains 获取所有域名列表
func (this *CustomHTTPProvider) GetDomains() (domains []string, err error) {
	resp, err := this.post(maps.Map{})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &domains)
	return
}

// GetRecords 获取域名解析记录列表
func (this *CustomHTTPProvider) GetRecords(domain string) (records []*dnstypes.Record, err error) {
	resp, err := this.post(maps.Map{
		"action": "GetRecords",
		"domain": domain,
	})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &records)
	return
}

// GetRoutes 读取域名支持的线路数据
func (this *CustomHTTPProvider) GetRoutes(domain string) (routes []*dnstypes.Route, err error) {
	resp, err := this.post(maps.Map{
		"action": "GetRoutes",
		"domain": domain,
	})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &routes)
	return
}

// QueryRecord 查询单个记录
func (this *CustomHTTPProvider) QueryRecord(domain string, name string, recordType dnstypes.RecordType) (*dnstypes.Record, error) {
	resp, err := this.post(maps.Map{
		"action":     "QueryRecord",
		"domain":     domain,
		"name":       name,
		"recordType": recordType,
	})
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, nil
	}
	record := &dnstypes.Record{}
	err = json.Unmarshal(resp, record)
	if err != nil {
		return nil, err
	}
	if len(record.Value) == 0 {
		return nil, nil
	}
	return record, nil
}

// AddRecord 设置记录
func (this *CustomHTTPProvider) AddRecord(domain string, newRecord *dnstypes.Record) error {
	_, err := this.post(maps.Map{
		"action":    "AddRecord",
		"domain":    domain,
		"newRecord": newRecord,
	})
	return err
}

// UpdateRecord 修改记录
func (this *CustomHTTPProvider) UpdateRecord(domain string, record *dnstypes.Record, newRecord *dnstypes.Record) error {
	_, err := this.post(maps.Map{
		"action":    "UpdateRecord",
		"domain":    domain,
		"record":    record,
		"newRecord": newRecord,
	})
	return err
}

// DeleteRecord 删除记录
func (this *CustomHTTPProvider) DeleteRecord(domain string, record *dnstypes.Record) error {
	_, err := this.post(maps.Map{
		"action": "DeleteRecord",
		"domain": domain,
		"record": record,
	})
	return err
}

// DefaultRoute 默认线路
func (this *CustomHTTPProvider) DefaultRoute() string {
	resp, err := this.post(maps.Map{
		"action": "DefaultRoute",
	})
	if err != nil {
		return ""
	}
	return string(resp)
}

// 执行操作
func (this *CustomHTTPProvider) post(params maps.Map) (respData []byte, err error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, this.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Set("Timestamp", timestamp)
	req.Header.Set("Token", fmt.Sprintf("%x", sha1.Sum([]byte(this.secret+"@"+timestamp))))
	req.Header.Set("User-Agent", "GoEdge/"+teaconst.Version)

	resp, err := customHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return nil, errors.New("status should be 200, but got '" + strconv.Itoa(resp.StatusCode) + "'")
	}
	return ioutil.ReadAll(resp.Body)
}
