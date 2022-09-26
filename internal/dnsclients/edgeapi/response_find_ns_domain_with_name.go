// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package edgeapi

type FindDomainWithNameResponse struct {
	BaseResponse

	Data struct {
		NSDomain struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		}
	} `json:"data"`
}
