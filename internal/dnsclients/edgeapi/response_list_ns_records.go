// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package edgeapi

type ListNSRecordsResponse struct {
	BaseResponse

	Data struct {
		NSRecords []struct {
			Id       int64  `json:"id"`
			Name     string `json:"name"`
			Value    string `json:"value"`
			TTL      int32  `json:"ttl"`
			Type     string `json:"type"`
			NSRoutes []struct {
				Name string `json:"name"`
				Code string `json:"code"`
			} `json:"nsRoutes"`
		} `json:"nsRecords"`
	} `json:"data"`
}
