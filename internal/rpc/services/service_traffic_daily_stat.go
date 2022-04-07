// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

// TrafficDailyStatService 按日统计服务
type TrafficDailyStatService struct {
	BaseService
}

// FindTrafficDailyStatWithDay 查找某日统计
func (this *TrafficDailyStatService) FindTrafficDailyStatWithDay(ctx context.Context, req *pb.FindTrafficDailyStatWithDayRequest) (*pb.FindTrafficDailyStatWithDayResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var day = timeutil.Format("Ymd")
	stat, err := stats.SharedTrafficDailyStatDAO.FindDailyStat(tx, day)
	if err != nil {
		return nil, err
	}

	if stat == nil {
		return &pb.FindTrafficDailyStatWithDayResponse{
			TrafficDailyStat: nil,
		}, nil
	}

	return &pb.FindTrafficDailyStatWithDayResponse{
		TrafficDailyStat: &pb.TrafficDailyStat{
			Id:                  int64(stat.Id),
			Day:                 stat.Day,
			CachedBytes:         int64(stat.CachedBytes),
			Bytes:               int64(stat.Bytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		},
	}, nil
}
