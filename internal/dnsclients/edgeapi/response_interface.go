// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package edgeapi

type ResponseInterface interface {
	IsValid() bool
	Error() error
}
