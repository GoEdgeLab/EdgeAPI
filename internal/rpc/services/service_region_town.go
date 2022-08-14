// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// RegionTownService 区县相关服务
type RegionTownService struct {
	BaseService
}

// FindAllRegionTowns 查找所有区县
func (this *RegionTownService) FindAllRegionTowns(ctx context.Context, req *pb.FindAllRegionTownsRequest) (*pb.FindAllRegionTownsResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	towns, err := regions.SharedRegionTownDAO.FindAllRegionTowns(tx)
	if err != nil {
		return nil, err
	}

	var pbTowns = []*pb.RegionTown{}

	var citiesMap = map[int64]*regions.RegionCity{} // provinceId => *RegionCity

	for _, town := range towns {
		var cityId = int64(town.CityId)

		var pbCity = &pb.RegionCity{Id: cityId}
		if req.IncludeRegionCity {
			city, ok := citiesMap[cityId]
			if !ok {
				city, err = regions.SharedRegionCityDAO.FindEnabledRegionCity(tx, cityId)
				if err != nil {
					return nil, err
				}
				if city == nil {
					continue
				}
				citiesMap[cityId] = city
			}

			pbCity = &pb.RegionCity{
				Id:          int64(city.Id),
				Name:        city.Name,
				Codes:       city.DecodeCodes(),
				CustomName:  city.CustomName,
				CustomCodes: city.DecodeCustomCodes(),
				DisplayName: city.DisplayName(),
			}
		}

		pbTowns = append(pbTowns, &pb.RegionTown{
			Id:           int64(town.Id),
			Name:         town.Name,
			Codes:        town.DecodeCodes(),
			RegionCityId: int64(town.CityId),
			RegionCity:   pbCity,
			CustomName:   town.CustomName,
			CustomCodes:  town.DecodeCustomCodes(),
			DisplayName:  town.DisplayName(),
		})
	}

	return &pb.FindAllRegionTownsResponse{
		RegionTowns: pbTowns,
	}, nil
}

// FindAllRegionTownsWithRegionCityId 查找某个城市的所有区县
func (this *RegionTownService) FindAllRegionTownsWithRegionCityId(ctx context.Context, req *pb.FindAllRegionTownsWithRegionCityIdRequest) (*pb.FindAllRegionTownsWithRegionCityIdResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	towns, err := regions.SharedRegionTownDAO.FindAllRegionTownsWithCityId(tx, req.RegionCityId)
	if err != nil {
		return nil, err
	}

	var pbTowns = []*pb.RegionTown{}

	for _, town := range towns {
		var cityId = int64(town.CityId)

		var pbCity = &pb.RegionCity{Id: cityId}

		pbTowns = append(pbTowns, &pb.RegionTown{
			Id:           int64(town.Id),
			Name:         town.Name,
			Codes:        town.DecodeCodes(),
			RegionCityId: int64(town.CityId),
			RegionCity:   pbCity,
			CustomName:   town.CustomName,
			CustomCodes:  town.DecodeCustomCodes(),
			DisplayName:  town.DisplayName(),
		})
	}

	return &pb.FindAllRegionTownsWithRegionCityIdResponse{
		RegionTowns: pbTowns,
	}, nil
}

// FindRegionTown 查找单个区县信息
func (this *RegionTownService) FindRegionTown(ctx context.Context, req *pb.FindRegionTownRequest) (*pb.FindRegionTownResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	town, err := regions.SharedRegionTownDAO.FindEnabledRegionTown(tx, req.RegionTownId)
	if err != nil {
		return nil, err
	}
	if town == nil {
		return &pb.FindRegionTownResponse{
			RegionTown: nil,
		}, nil
	}

	return &pb.FindRegionTownResponse{
		RegionTown: &pb.RegionTown{
			Id:           int64(town.Id),
			Name:         town.Name,
			Codes:        town.DecodeCodes(),
			RegionCityId: int64(town.CityId),
			CustomName:   town.CustomName,
			CustomCodes:  town.DecodeCustomCodes(),
			DisplayName:  town.DisplayName(),
		},
	}, nil
}

// UpdateRegionTownCustom 修改区县定制信息
func (this *RegionTownService) UpdateRegionTownCustom(ctx context.Context, req *pb.UpdateRegionTownCustomRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = regions.SharedRegionTownDAO.UpdateTownCustom(tx, req.RegionTownId, req.CustomName, req.CustomCodes)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
