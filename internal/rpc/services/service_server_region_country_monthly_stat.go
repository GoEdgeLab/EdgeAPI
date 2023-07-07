package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 地区月份统计
type ServerRegionCountryMonthlyStatService struct {
	BaseService
}

// 查找前N个地区
func (this *ServerRegionCountryMonthlyStatService) FindTopServerRegionCountryMonthlyStats(ctx context.Context, req *pb.FindTopServerRegionCountryMonthlyStatsRequest) (*pb.FindTopServerRegionCountryMonthlyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()
	statList, err := stats.SharedServerRegionCountryMonthlyStatDAO.ListStats(tx, req.ServerId, req.Month, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbStats = []*pb.FindTopServerRegionCountryMonthlyStatsResponse_Stat{}
	for _, stat := range statList {
		pbStat := &pb.FindTopServerRegionCountryMonthlyStatsResponse_Stat{
			Count: int64(stat.Count),
		}

		country, err := regions.SharedRegionCountryDAO.FindEnabledRegionCountry(tx, int64(stat.CountryId))
		if err != nil {
			return nil, err
		}
		if country == nil {
			continue
		}
		pbStat.RegionCountry = &pb.RegionCountry{
			Id:   int64(country.ValueId),
			Name: country.DisplayName(),
		}

		pbStats = append(pbStats, pbStat)
	}

	return &pb.FindTopServerRegionCountryMonthlyStatsResponse{Stats: pbStats}, nil
}
