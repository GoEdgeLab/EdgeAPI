// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build community
// +build community

package authority

// IsPlus 判断是否为企业版
func (this *AuthorityKeyDAO) IsPlus(tx *dbs.Tx) (bool, error) {
	return false
}
