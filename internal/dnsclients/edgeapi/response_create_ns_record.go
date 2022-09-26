// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package edgeapi

type CreateNSRecordResponse struct {
	BaseResponse

	Data struct {
		NSRecordId int64 `json:"nsRecordId"`
	} `json:"data"`
}
