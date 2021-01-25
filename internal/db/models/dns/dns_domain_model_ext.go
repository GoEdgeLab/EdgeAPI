package dns

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
)

// 获取所有的线路
func (this *DNSDomain) DecodeRoutes() ([]*dnsclients.Route, error) {
	if len(this.Routes) == 0 || this.Routes == "null" {
		return nil, nil
	}
	result := []*dnsclients.Route{}
	err := json.Unmarshal([]byte(this.Routes), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 检查是否包含某个线路
func (this *DNSDomain) ContainsRouteCode(route string) (bool, error) {
	routes, err := this.DecodeRoutes()
	if err != nil {
		return false, err
	}
	for _, r := range routes {
		if r.Code == route {
			return true, nil
		}
	}
	return false, nil
}

// 获取所有的记录
func (this *DNSDomain) DecodeRecords() ([]*dnsclients.Record, error) {
	records := this.Records
	if len(records) == 0 || records == "null" {
		return nil, nil
	}
	result := []*dnsclients.Record{}
	err := json.Unmarshal([]byte(records), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
