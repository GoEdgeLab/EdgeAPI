// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnspod

type BaseResponse struct {
	Status struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
}

func (this *BaseResponse) IsOk() bool {
	return this.Status.Code == "1"
}

func (this *BaseResponse) LastError() (code string, message string) {
	return this.Status.Code, this.Status.Message
}
