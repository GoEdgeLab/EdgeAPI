// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import "github.com/iwind/TeaGo/dbs"

func (this *UserPlanStatDAO) IncreaseUserPlanStat(tx *dbs.Tx, userPlanId int64, trafficBytes int64, countRequests int64) error {
	return nil
}
