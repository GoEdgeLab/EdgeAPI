// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnspod

type DomainInfoResponse struct {
	BaseResponse

	Domain struct {
		Id    any    `json:"id"`
		Name  string `json:"name"`
		Grade string `json:"grade"`
	} `json:"domain"`
}
