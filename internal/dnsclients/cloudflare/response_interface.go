// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package cloudflare

type ResponseInterface interface {
	IsOk() bool
	LastError() (code int, message string)
}
