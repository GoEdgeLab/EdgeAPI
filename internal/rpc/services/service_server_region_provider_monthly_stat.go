package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 运营商月份统计
type ServerRegionProviderMonthlyStatService struct {
	BaseService
}

// 查找前N个运营商
func (this *ServerRegionProviderMonthlyStatService) FindTopServerRegionProviderMonthlyStats(ctx context.Context, req *pb.FindTopServerRegionProviderMonthlyStatsRequest) (*pb.FindTopServerRegionProviderMonthlyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
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
	statList, err := stats.SharedServerRegionProviderMonthlyStatDAO.ListStats(tx, req.ServerId, req.Month, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	pbStats := []*pb.FindTopServerRegionProviderMonthlyStatsResponse_Stat{}
	for _, stat := range statList {
		pbStat := &pb.FindTopServerRegionProviderMonthlyStatsResponse_Stat{
			Count: int64(stat.Count),
		}
		provider, err := regions.SharedRegionProviderDAO.FindEnabledRegionProvider(tx, stat.ProviderId)
		if err != nil {
			return nil, err
		}
		if provider == nil {
			continue
		}
		pbStat.RegionProvider = &pb.RegionProvider{
			Id:   int64(provider.Id),
			Name: provider.Name,
		}
		pbStats = append(pbStats, pbStat)
	}
	return &pb.FindTopServerRegionProviderMonthlyStatsResponse{Stats: pbStats}, nil
}
