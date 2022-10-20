// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnspod

type ResponseInterface interface {
	IsOk() bool
	LastError() (code string, message string)
}
