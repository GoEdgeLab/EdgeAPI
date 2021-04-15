// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package cloudflare

type ZonesResponse struct {
	BaseResponse

	Result []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}
