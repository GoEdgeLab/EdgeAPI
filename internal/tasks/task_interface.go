// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package tasks

type TaskInterface interface {
	Start() error
	Loop() error
	Stop() error
}
