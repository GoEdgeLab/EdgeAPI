// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build !plus
// +build !plus

package authority

import (
	"github.com/iwind/TeaGo/dbs"
)

// IsPlus 判断是否为企业版
func (this *AuthorityKeyDAO) IsPlus(tx *dbs.Tx) (bool, error) {
	return false, nil
}
