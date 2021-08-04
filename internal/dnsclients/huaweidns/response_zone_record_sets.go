// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package huaweidns

type ZoneRecordSetsResponse struct {
	RecordSets []struct {
		Id      string   `json:"id"`
		Name    string   `json:"name"`
		Type    string   `json:"type"`
		Ttl     int      `json:"ttl"`
		Records []string `json:"records"`
		Line    string   `json:"line"`
	} `json:"recordsets"`
	Metadata struct {
		TotalCount int `json:"total_count"`
	} `json:"metadata"`
}
