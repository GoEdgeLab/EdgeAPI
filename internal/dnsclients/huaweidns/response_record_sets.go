// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package huaweidns

type RecordSetsResponse struct {
	RecordSets []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		Ttl  int    `json:"ttl"`
		Line string `json:"line"`
		Records []string `json:"records"`
	} `json:"recordsets"`
}
