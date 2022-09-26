// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package edgeapi

type ListNSDomainsResponse struct {
	BaseResponse

	Data struct {
		NSDomains []struct {
			Id        int64  `json:"id"`
			Name      string `json:"name"`
			IsOn      bool   `json:"isOn"`
			IsDeleted bool   `json:"isDeleted"`
		} `json:"nsDomains"`
	} `json:"data"`
}
