// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package huaweidns

type ZonesResponse struct {
	Links struct{} `json:"links"`
	Zones []struct {
		Id        string `json:"id"`
		Name      string `json:"name"`
		ZoneType  string `json:"zone_type"`
		Status    string `json:"status"`
		RecordNum int    `json:"record_num"`
	} `json:"zones"`
	Metadata struct {
		TotalCount int `json:"total_count"`
	} `json:"metadata"`
}
