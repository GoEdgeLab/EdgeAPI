// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
	"regexp"
)

// TrafficDailyStatService 按日统计服务
type TrafficDailyStatService struct {
	BaseService
}

// FindTrafficDailyStatWithDay 查找某日统计
func (this *TrafficDailyStatService) FindTrafficDailyStatWithDay(ctx context.Context, req *pb.FindTrafficDailyStatWithDayRequest) (*pb.FindTrafficDailyStatWithDayResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var day = req.Day
	var stat = &stats.TrafficDailyStat{
		Day: day,
	}
	if len(req.Minute) > 0 && regexp.MustCompile(`^\d{6}$`).MatchString(req.Minute) {
		var hourString = req.Minute[:2]
		var hourInt = types.Int(hourString)
		var lastHourInt = hourInt - 1

		// 过往小时
		if lastHourInt >= 0 {
			var hourFrom = day + "00"
			var hourTo = day + fmt.Sprintf("%02d", lastHourInt)
			sumStat, err := stats.SharedTrafficHourlyStatDAO.SumHourlyStats(tx, hourFrom, hourTo)
			if err != nil {
				return nil, err
			}
			if sumStat != nil {
				stat = &stats.TrafficDailyStat{
					Id:                  0,
					Day:                 day,
					CachedBytes:         sumStat.CachedBytes,
					Bytes:               sumStat.Bytes,
					CountRequests:       sumStat.CountRequests,
					CountCachedRequests: sumStat.CountCachedRequests,
					CountAttackRequests: sumStat.CountAttackRequests,
					AttackBytes:         sumStat.AttackBytes,
				}
			}
		}

		// 当前小时
		hourStat, err := stats.SharedTrafficHourlyStatDAO.FindHourlyStat(tx, day+hourString)
		if err != nil {
			return nil, err
		}
		if hourStat != nil {
			var minutes = types.Int(req.Minute[2:4])/10 + 1 // 这里使用10分钟（所以下面是6），而不是秒，以便于让数据不至于变化太大，毕竟流量不是事实统计的
			stat.Bytes += hourStat.Bytes * uint64(minutes) / 6
			stat.CachedBytes += hourStat.CachedBytes * uint64(minutes) / 6
			stat.CountRequests += hourStat.CountRequests * uint64(minutes) / 6
			stat.CountCachedRequests += hourStat.CountCachedRequests * uint64(minutes) / 6
			stat.CountAttackRequests += hourStat.CountAttackRequests * uint64(minutes) / 6
			stat.AttackBytes += hourStat.AttackBytes * uint64(minutes) / 6
		}
	} else {
		stat, err = stats.SharedTrafficDailyStatDAO.FindDailyStat(tx, day)
		if err != nil {
			return nil, err
		}
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
