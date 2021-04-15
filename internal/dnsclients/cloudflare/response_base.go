// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package cloudflare

type BaseResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func (this *BaseResponse) IsOk() bool {
	return this.Success
}

func (this *BaseResponse) LastError() (code int, message string) {
	if len(this.Errors) == 0 {
		return 0, ""
	}
	return this.Errors[0].Code, this.Errors[0].Message
}
