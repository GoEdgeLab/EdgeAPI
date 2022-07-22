// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

// APIMethodStatService API方法统计服务
type APIMethodStatService struct {
	BaseService
}

// FindAPIMethodStatsWithDay 查找某天的统计
func (this *APIMethodStatService) FindAPIMethodStatsWithDay(ctx context.Context, req *pb.FindAPIMethodStatsWithDayRequest) (*pb.FindAPIMethodStatsWithDayResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var day = req.Day
	if len(day) == 0 {
		day = timeutil.Format("Ymd")
	}
	var tx = this.NullTx()
	stats, err := models.SharedAPIMethodStatDAO.FindAllStatsWithDay(tx, day)
	if err != nil {
		return nil, err
	}
	var pbStats = []*pb.APIMethodStat{}
	var cacheMap = utils.NewCacheMap()
	for _, stat := range stats {
		apiNode, err := models.SharedAPINodeDAO.FindEnabledAPINode(tx, int64(stat.ApiNodeId), cacheMap)
		if err != nil {
			return nil, err
		}
		if apiNode == nil {
			continue
		}

		pbStats = append(pbStats, &pb.APIMethodStat{
			Id:         int64(stat.Id),
			ApiNodeId:  int64(stat.ApiNodeId),
			Method:     stat.Method,
			Tag:        stat.Tag,
			CostMs:     float32(stat.CostMs),
			PeekMs:     float32(stat.PeekMs),
			CountCalls: int64(stat.CountCalls),
			ApiNode: &pb.APINode{
				Id:   int64(apiNode.Id),
				Name: apiNode.Name,
			},
		})
	}

	return &pb.FindAPIMethodStatsWithDayResponse{
		ApiMethodStats: pbStats,
	}, nil
}

// CountAPIMethodStatsWithDay 检查是否有统计数据
func (this *APIMethodStatService) CountAPIMethodStatsWithDay(ctx context.Context, req *pb.CountAPIMethodStatsWithDayRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var day = req.Day
	if len(day) == 0 {
		day = timeutil.Format("Ymd")
	}

	var tx = this.NullTx()
	count, err := models.SharedAPIMethodStatDAO.CountAllStatsWithDay(tx, day)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}
