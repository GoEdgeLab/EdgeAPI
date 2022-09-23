// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package setup

import (
	"github.com/iwind/TeaGo/dbs"
)

// 检查自建DNS全局设置
func (this *SQLExecutor) checkNS(db *dbs.DB) error {
	return nil
}
