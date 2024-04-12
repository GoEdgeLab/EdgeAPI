// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import "github.com/iwind/TeaGo/dbs"

// ResetServersTrafficLimitStatusWithUserPlanId 重置用户套餐相关网站限流状态
func (this *ServerDAO) ResetServersTrafficLimitStatusWithUserPlanId(tx *dbs.Tx, userPlanId int64) error {
	// stub
	return nil
}
