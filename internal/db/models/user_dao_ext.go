// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import "github.com/iwind/TeaGo/dbs"

// CheckUserServersEnabled 判断用户是否可用服务功能
func (this *UserDAO) CheckUserServersEnabled(tx *dbs.Tx, userId int64) (isEnabled bool, err error) {
	// 是否已删除、未启用、已拒绝
	one, err := this.Query(tx).
		Result("id", "isRejected", "state", "isOn").
		Pk(userId).
		Find()
	if err != nil {
		return false, err
	}
	if one == nil {
		return false, nil
	}
	var user = one.(*User)
	if user.State != UserStateEnabled || !user.IsOn || user.IsRejected {
		return false, nil
	}

	return true, nil
}
