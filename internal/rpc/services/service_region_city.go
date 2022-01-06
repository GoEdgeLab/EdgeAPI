// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// RegionCityService 城市相关服务
type RegionCityService struct {
	BaseService
}

// FindAllEnabledRegionCities 查找所有城市
func (this *RegionCityService) FindAllEnabledRegionCities(ctx context.Context, req *pb.FindAllEnabledRegionCitiesRequest) (*pb.FindAllEnabledRegionCitiesResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	cities, err := regions.SharedRegionCityDAO.FindAllEnabledCities(tx)
	if err != nil {
		return nil, err
	}

	var pbCities = []*pb.RegionCity{}
	for _, city := range cities {
		pbCities = append(pbCities, &pb.RegionCity{
			Id:               int64(city.Id),
			Name:             city.Name,
			Codes:            city.DecodeCodes(),
			RegionProvinceId: int64(city.ProvinceId),
		})
	}

	return &pb.FindAllEnabledRegionCitiesResponse{
		RegionCities: pbCities,
	}, nil
}

// FindEnabledRegionCity 查找单个城市信息
func (this *RegionCityService) FindEnabledRegionCity(ctx context.Context, req *pb.FindEnabledRegionCityRequest) (*pb.FindEnabledRegionCityResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	city, err := regions.SharedRegionCityDAO.FindEnabledRegionCity(tx, req.RegionCityId)
	if err != nil {
		return nil, err
	}
	if city == nil {
		return &pb.FindEnabledRegionCityResponse{
			RegionCity: nil,
		}, nil
	}

	return &pb.FindEnabledRegionCityResponse{
		RegionCity: &pb.RegionCity{
			Id:               int64(city.Id),
			Name:             city.Name,
			Codes:            city.DecodeCodes(),
			RegionProvinceId: int64(city.ProvinceId),
		},
	}, nil
}
