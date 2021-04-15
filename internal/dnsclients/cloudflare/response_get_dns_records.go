// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package cloudflare

type GetDNSRecordsResponse struct {
	BaseResponse

	Result []struct {
		Id       string `json:"id"`
		Type     string `json:"type"`
		Name     string `json:"name"`
		Content  string `json:"content"`
		Ttl      int    `json:"ttl"`
		ZoneId   string `json:"zoneId"`
		ZoneName string `json:"zoneName"`
	} `json:"result"`
}
