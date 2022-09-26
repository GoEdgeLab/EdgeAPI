// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package edgeapi

import (
	"errors"
	"github.com/iwind/TeaGo/types"
)

type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (this *BaseResponse) IsValid() bool {
	return this.Code == 200
}

func (this *BaseResponse) Error() error {
	return errors.New("code: " + types.String(this.Code) + ", message: " + this.Message)
}
