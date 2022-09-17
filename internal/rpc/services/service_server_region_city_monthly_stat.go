package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 城市月份统计
type ServerRegionCityMonthlyStatService struct {
	BaseService
}

// 查找前N个城市
func (this *ServerRegionCityMonthlyStatService) FindTopServerRegionCityMonthlyStats(ctx context.Context, req *pb.FindTopServerRegionCityMonthlyStatsRequest) (*pb.FindTopServerRegionCityMonthlyStatsResponse, error) {
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
	statList, err := stats.SharedServerRegionCityMonthlyStatDAO.ListStats(tx, req.ServerId, req.Month, req.CountryId, req.ProvinceId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbStats = []*pb.FindTopServerRegionCityMonthlyStatsResponse_Stat{}
	for _, stat := range statList {
		pbStat := &pb.FindTopServerRegionCityMonthlyStatsResponse_Stat{
			Count: int64(stat.Count),
		}
		if stat.CityId == 0 {
			continue
		}
		city, err := regions.SharedRegionCityDAO.FindEnabledRegionCity(tx, int64(stat.CityId))
		if err != nil {
			return nil, err
		}
		if city == nil {
			continue
		}
		province, err := regions.SharedRegionProvinceDAO.FindEnabledRegionProvince(tx, int64(city.ProvinceId))
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
		pbStat.RegionCity = &pb.RegionCity{
			Id:   int64(city.Id),
			Name: city.DisplayName(),
		}
		pbStats = append(pbStats, pbStat)
	}
	return &pb.FindTopServerRegionCityMonthlyStatsResponse{Stats: pbStats}, nil
}
