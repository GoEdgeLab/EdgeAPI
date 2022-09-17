package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 省份月份统计
type ServerRegionProvinceMonthlyStatService struct {
	BaseService
}

// 查找前N个省份
func (this *ServerRegionProvinceMonthlyStatService) FindTopServerRegionProvinceMonthlyStats(ctx context.Context, req *pb.FindTopServerRegionProvinceMonthlyStatsRequest) (*pb.FindTopServerRegionProvinceMonthlyStatsResponse, error) {
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
	statList, err := stats.SharedServerRegionProvinceMonthlyStatDAO.ListStats(tx, req.ServerId, req.Month, req.CountryId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbStats = []*pb.FindTopServerRegionProvinceMonthlyStatsResponse_Stat{}
	for _, stat := range statList {
		pbStat := &pb.FindTopServerRegionProvinceMonthlyStatsResponse_Stat{
			Count: int64(stat.Count),
		}
		province, err := regions.SharedRegionProvinceDAO.FindEnabledRegionProvince(tx, int64(stat.ProvinceId))
		if err != nil {
			return nil, err
		}
		if province == nil {
			continue
		}
		country, err := regions.SharedRegionCountryDAO.FindEnabledRegionCountry(tx, int64(province.CountryId))
		if err != nil {
			return nil, err
		}
		if country == nil {
			continue
		}
		pbStat.RegionCountry = &pb.RegionCountry{
			Id:   int64(country.Id),
			Name: country.DisplayName(),
		}
		pbStat.RegionProvince = &pb.RegionProvince{
			Id:   int64(province.Id),
			Name: province.DisplayName(),
		}
		pbStats = append(pbStats, pbStat)
	}
	return &pb.FindTopServerRegionProvinceMonthlyStatsResponse{Stats: pbStats}, nil
}
