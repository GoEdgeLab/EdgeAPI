package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/iwind/TeaGo/lists"
)

// 获取所有的线路
func (this *DNSDomain) DecodeRoutes() ([]string, error) {
	if len(this.Routes) == 0 || this.Routes == "null" {
		return nil, nil
	}
	result := []string{}
	err := json.Unmarshal([]byte(this.Routes), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 检查是否包含某个线路
func (this *DNSDomain) ContainsRoute(route string) (bool, error) {
	routes, err := this.DecodeRoutes()
	if err != nil {
		return false, err
	}
	return lists.ContainsString(routes, route), nil
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
