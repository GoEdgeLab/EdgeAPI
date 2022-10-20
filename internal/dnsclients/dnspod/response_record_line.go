// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnspod

type RecordLineResponse struct {
	BaseResponse

	Lines []string `json:"lines"`
}
