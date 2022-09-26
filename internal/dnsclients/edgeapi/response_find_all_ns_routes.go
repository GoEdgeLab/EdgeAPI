// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package edgeapi

type FindAllNSRoutesResponse struct {
	BaseResponse

	Data struct {
		NSRoutes []struct {
			Name string `json:"name"`
			Code string `json:"code"`
		} `json:"nsRoutes"`
	} `json:"data"`
}
