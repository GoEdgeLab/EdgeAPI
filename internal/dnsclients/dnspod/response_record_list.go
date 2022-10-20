// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnspod

type RecordListResponse struct {
	BaseResponse

	Info struct {
		SubDomains  string `json:"sub_domains"`
		RecordTotal string `json:"record_total"`
		RecordsNum  string `json:"records_num"`
	} `json:"info"`

	Records []struct {
		Id     any    `json:"id"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Value  string `json:"value"`
		Line   string `json:"line"`
		LineId string `json:"line_id"`
		TTL    string `json:"ttl"`
	} `json:"records"`
}
