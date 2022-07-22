// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ServerDomainHourlyStatService 服务域名按小时统计服务
type ServerDomainHourlyStatService struct {
	BaseService
}

// ListTopServerDomainStatsWithServerId 读取域名排行
func (this *ServerDomainHourlyStatService) ListTopServerDomainStatsWithServerId(ctx context.Context, req *pb.ListTopServerDomainStatsWithServerIdRequest) (*pb.ListTopServerDomainStatsWithServerIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var topDomainStats []*stats.ServerDomainHourlyStat
	if req.ServerId > 0 {
		topDomainStats, err = stats.SharedServerDomainHourlyStatDAO.FindTopDomainStatsWithServerId(tx, req.ServerId, req.HourFrom, req.HourTo, req.Size)
	} else if req.NodeId > 0 {
		// 域名排行
		topDomainStats, err = stats.SharedServerDomainHourlyStatDAO.FindTopDomainStatsWithNodeId(tx, req.NodeId, req.HourFrom, req.HourTo, 10)
	} else if req.NodeClusterId > 0 {
		topDomainStats, err = stats.SharedServerDomainHourlyStatDAO.FindTopDomainStatsWithClusterId(tx, req.NodeClusterId, req.HourFrom, req.HourTo, 10)
	}
	if err != nil {
		return nil, err
	}

	var pbDomainStats = []*pb.ServerDomainHourlyStat{}
	for _, stat := range topDomainStats {
		pbDomainStats = append(pbDomainStats, &pb.ServerDomainHourlyStat{
			ServerId:            int64(stat.ServerId),
			Domain:              stat.Domain,
			CountRequests:       int64(stat.CountRequests),
			Bytes:               int64(stat.Bytes),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}

	return &pb.ListTopServerDomainStatsWithServerIdResponse{
		DomainStats: pbDomainStats,
	}, nil
}
