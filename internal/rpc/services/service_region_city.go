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
// Deprecated
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

	var provincesMap = map[int64]*regions.RegionProvince{} // provinceId => RegionProvince

	for _, city := range cities {
		var provinceId = int64(city.ProvinceId)

		var pbProvince = &pb.RegionProvince{Id: provinceId}
		if req.IncludeRegionProvince {
			province, ok := provincesMap[provinceId]
			if !ok {
				province, err = regions.SharedRegionProvinceDAO.FindEnabledRegionProvince(tx, provinceId)
				if err != nil {
					return nil, err
				}
				if province == nil {
					continue
				}
				provincesMap[provinceId] = province
			}

			pbProvince = &pb.RegionProvince{
				Id:          int64(province.ValueId),
				Name:        province.Name,
				Codes:       province.DecodeCodes(),
				DisplayName: province.DisplayName(),
			}
		}

		pbCities = append(pbCities, &pb.RegionCity{
			Id:               int64(city.ValueId),
			Name:             city.Name,
			Codes:            city.DecodeCodes(),
			RegionProvinceId: int64(city.ProvinceId),
			RegionProvince:   pbProvince,
			CustomName:       city.CustomName,
			CustomCodes:      city.DecodeCustomCodes(),
			DisplayName:      city.DisplayName(),
		})
	}

	return &pb.FindAllEnabledRegionCitiesResponse{
		RegionCities: pbCities,
	}, nil
}

// FindAllRegionCities 查找所有城市
func (this *RegionCityService) FindAllRegionCities(ctx context.Context, req *pb.FindAllRegionCitiesRequest) (*pb.FindAllRegionCitiesResponse, error) {
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

	var provincesMap = map[int64]*regions.RegionProvince{} // provinceId => RegionProvince

	for _, city := range cities {
		var provinceId = int64(city.ProvinceId)

		var pbProvince = &pb.RegionProvince{Id: provinceId}
		if req.IncludeRegionProvince {
			province, ok := provincesMap[provinceId]
			if !ok {
				province, err = regions.SharedRegionProvinceDAO.FindEnabledRegionProvince(tx, provinceId)
				if err != nil {
					return nil, err
				}
				if province == nil {
					continue
				}
				provincesMap[provinceId] = province
			}

			pbProvince = &pb.RegionProvince{
				Id:          int64(province.ValueId),
				Name:        province.Name,
				Codes:       province.DecodeCodes(),
				CustomName:  province.CustomName,
				CustomCodes: province.DecodeCustomCodes(),
				DisplayName: province.DisplayName(),
			}
		}

		pbCities = append(pbCities, &pb.RegionCity{
			Id:               int64(city.ValueId),
			Name:             city.Name,
			Codes:            city.DecodeCodes(),
			RegionProvinceId: int64(city.ProvinceId),
			RegionProvince:   pbProvince,
			CustomName:       city.CustomName,
			CustomCodes:      city.DecodeCustomCodes(),
			DisplayName:      city.DisplayName(),
		})
	}

	return &pb.FindAllRegionCitiesResponse{
		RegionCities: pbCities,
	}, nil
}

// FindAllRegionCitiesWithRegionProvinceId 查找某个省份的所有城市
func (this *RegionCityService) FindAllRegionCitiesWithRegionProvinceId(ctx context.Context, req *pb.FindAllRegionCitiesWithRegionProvinceIdRequest) (*pb.FindAllRegionCitiesWithRegionProvinceIdResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	cities, err := regions.SharedRegionCityDAO.FindAllEnabledCitiesWithProvinceId(tx, req.RegionProvinceId)
	if err != nil {
		return nil, err
	}

	var pbCities = []*pb.RegionCity{}

	for _, city := range cities {
		var provinceId = int64(city.ProvinceId)

		var pbProvince = &pb.RegionProvince{Id: provinceId}

		pbCities = append(pbCities, &pb.RegionCity{
			Id:               int64(city.ValueId),
			Name:             city.Name,
			Codes:            city.DecodeCodes(),
			RegionProvinceId: int64(city.ProvinceId),
			RegionProvince:   pbProvince,
			CustomName:       city.CustomName,
			CustomCodes:      city.DecodeCustomCodes(),
			DisplayName:      city.DisplayName(),
		})
	}

	return &pb.FindAllRegionCitiesWithRegionProvinceIdResponse{
		RegionCities: pbCities,
	}, nil
}

// FindEnabledRegionCity 查找单个城市信息
// Deprecated
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
			Id:               int64(city.ValueId),
			Name:             city.Name,
			Codes:            city.DecodeCodes(),
			RegionProvinceId: int64(city.ProvinceId),
			CustomName:       city.CustomName,
			CustomCodes:      city.DecodeCustomCodes(),
			DisplayName:      city.DisplayName(),
		},
	}, nil
}

// FindRegionCity 查找单个城市信息
func (this *RegionCityService) FindRegionCity(ctx context.Context, req *pb.FindRegionCityRequest) (*pb.FindRegionCityResponse, error) {
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
		return &pb.FindRegionCityResponse{
			RegionCity: nil,
		}, nil
	}

	return &pb.FindRegionCityResponse{
		RegionCity: &pb.RegionCity{
			Id:               int64(city.ValueId),
			Name:             city.Name,
			Codes:            city.DecodeCodes(),
			RegionProvinceId: int64(city.ProvinceId),
			CustomName:       city.CustomName,
			CustomCodes:      city.DecodeCustomCodes(),
			DisplayName:      city.DisplayName(),
		},
	}, nil
}

// UpdateRegionCityCustom 修改城市定制信息
func (this *RegionCityService) UpdateRegionCityCustom(ctx context.Context, req *pb.UpdateRegionCityCustomRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = regions.SharedRegionCityDAO.UpdateCityCustom(tx, req.RegionCityId, req.CustomName, req.CustomCodes)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
