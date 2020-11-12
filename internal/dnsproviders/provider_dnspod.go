package dnsproviders

import (
	"encoding/json"
	"errors"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type DNSPodProvider struct {
	apiId    string
	apiToken string
}

// 认证
func (this *DNSPodProvider) Auth(params maps.Map) error {
	this.apiId = params.GetString("id")
	this.apiToken = params.GetString("token")

	if len(this.apiId) == 0 {
		return errors.New("'id' should be not empty")
	}
	if len(this.apiToken) == 0 {
		return errors.New("'token' should not be empty")
	}
	return nil
}

// 读取线路数据
func (this *DNSPodProvider) GetRoutes(domain string) ([][]string, error) {
	infoResp, err := this.post("/Domain.info", map[string]string{
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
	lineStrings := []string{}
	for _, line := range lines {
		lineStrings = append(lineStrings, types.String(line))
	}

	return [][]string{lineStrings}, nil
}

// 发送请求
func (this *DNSPodProvider) post(path string, params map[string]string) (maps.Map, error) {
	apiHost := "https://dnsapi.cn"
	query := url.Values{
		"login_token": []string{this.apiId + "," + this.apiToken},
		"format":      []string{"json"},
	}
	for p, v := range params {
		query[p] = []string{v}
	}
	req, err := http.NewRequest(http.MethodPost, apiHost+path, strings.NewReader(query.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GoEdge Client/1.0.0 (iwind.liu@gmail.com)")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
		client.CloseIdleConnections()
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := maps.Map{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}
	status := m.GetMap("status")
	code := status.GetString("code")
	if code != "1" {
		return nil, errors.New("code: " + code + ", message: " + status.GetString("message"))
	}

	return m, nil
}
