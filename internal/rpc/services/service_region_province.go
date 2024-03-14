package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// RegionProvinceService 省份相关服务
type RegionProvinceService struct {
	BaseService
}

// FindAllEnabledRegionProvincesWithCountryId 查找所有省份
// Deprecated
func (this *RegionProvinceService) FindAllEnabledRegionProvincesWithCountryId(ctx context.Context, req *pb.FindAllEnabledRegionProvincesWithCountryIdRequest) (*pb.FindAllEnabledRegionProvincesWithCountryIdResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	provinces, err := regions.SharedRegionProvinceDAO.FindAllEnabledProvincesWithCountryId(tx, req.RegionCountryId)
	if err != nil {
		return nil, err
	}
	result := []*pb.RegionProvince{}
	for _, province := range provinces {
		result = append(result, &pb.RegionProvince{
			Id:          int64(province.ValueId),
			Name:        province.Name,
			Codes:       province.DecodeCodes(),
			CustomName:  province.CustomName,
			CustomCodes: province.DecodeCustomCodes(),
			DisplayName: province.DisplayName(),
		})
	}

	return &pb.FindAllEnabledRegionProvincesWithCountryIdResponse{
		RegionProvinces: result,
	}, nil
}

// FindEnabledRegionProvince 查找单个省份信息
// Deprecated
func (this *RegionProvinceService) FindEnabledRegionProvince(ctx context.Context, req *pb.FindEnabledRegionProvinceRequest) (*pb.FindEnabledRegionProvinceResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	province, err := regions.SharedRegionProvinceDAO.FindEnabledRegionProvince(tx, req.RegionProvinceId)
	if err != nil {
		return nil, err
	}
	if province == nil {
		return &pb.FindEnabledRegionProvinceResponse{RegionProvince: nil}, nil
	}

	return &pb.FindEnabledRegionProvinceResponse{
		RegionProvince: &pb.RegionProvince{
			Id:          int64(province.ValueId),
			Name:        province.Name,
			Codes:       province.DecodeCodes(),
			CustomName:  province.CustomName,
			CustomCodes: province.DecodeCustomCodes(),
			DisplayName: province.DisplayName(),
		},
	}, nil
}

// FindAllRegionProvincesWithRegionCountryId 查找所有省份
func (this *RegionProvinceService) FindAllRegionProvincesWithRegionCountryId(ctx context.Context, req *pb.FindAllRegionProvincesWithRegionCountryIdRequest) (*pb.FindAllRegionProvincesWithRegionCountryIdResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	provinces, err := regions.SharedRegionProvinceDAO.FindAllEnabledProvincesWithCountryId(tx, req.RegionCountryId)
	if err != nil {
		return nil, err
	}
	var pbProvinces = []*pb.RegionProvince{}
	for _, province := range provinces {
		pbProvinces = append(pbProvinces, &pb.RegionProvince{
			Id:          int64(province.ValueId),
			Name:        province.Name,
			Codes:       province.DecodeCodes(),
			CustomName:  province.CustomName,
			CustomCodes: province.DecodeCustomCodes(),
			DisplayName: province.DisplayName(),
		})
	}

	return &pb.FindAllRegionProvincesWithRegionCountryIdResponse{
		RegionProvinces: pbProvinces,
	}, nil
}

// FindAllRegionProvinces 查找所有国家|地区的所有省份
func (this *RegionProvinceService) FindAllRegionProvinces(ctx context.Context, req *pb.FindAllRegionProvincesRequest) (*pb.FindAllRegionProvincesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	provinces, err := regions.SharedRegionProvinceDAO.FindAllEnabledProvinces(tx)
	if err != nil {
		return nil, err
	}

	var pbProvinces = []*pb.RegionProvince{}
	var countryRouteCodeCache = map[int64]string{} // countryId => routeCode
	for _, province := range provinces {
		// 国家|地区ID
		var countryId = int64(province.CountryId)
		if countryId <= 0 {
			continue
		}

		// 国家|地区线路
		countryRouteCode, ok := countryRouteCodeCache[countryId]
		if !ok {
			countryRouteCode, err = regions.SharedRegionCountryDAO.FindRegionCountryRouteCode(tx, countryId)
			if err != nil {
				return nil, err
			}
			countryRouteCodeCache[countryId] = countryRouteCode
		}
		if len(countryRouteCode) == 0 {
			continue
		}

		pbProvinces = append(pbProvinces, &pb.RegionProvince{
			Id:              int64(province.ValueId),
			Name:            province.Name,
			Codes:           province.DecodeCodes(),
			CustomName:      province.CustomName,
			CustomCodes:     province.DecodeCustomCodes(),
			DisplayName:     province.DisplayName(),
			RegionCountryId: countryId,
			RegionCountry: &pb.RegionCountry{
				Id:        countryId,
				RouteCode: countryRouteCode,
			},
		})
	}
	return &pb.FindAllRegionProvincesResponse{
		RegionProvinces: pbProvinces,
	}, nil
}

// FindRegionProvince 查找单个省份信息
func (this *RegionProvinceService) FindRegionProvince(ctx context.Context, req *pb.FindRegionProvinceRequest) (*pb.FindRegionProvinceResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	province, err := regions.SharedRegionProvinceDAO.FindEnabledRegionProvince(tx, req.RegionProvinceId)
	if err != nil {
		return nil, err
	}
	if province == nil {
		return &pb.FindRegionProvinceResponse{RegionProvince: nil}, nil
	}

	return &pb.FindRegionProvinceResponse{
		RegionProvince: &pb.RegionProvince{
			Id:          int64(province.ValueId),
			Name:        province.Name,
			Codes:       province.DecodeCodes(),
			CustomName:  province.CustomName,
			CustomCodes: province.DecodeCustomCodes(),
			DisplayName: province.DisplayName(),
		},
	}, nil
}

// UpdateRegionProvinceCustom 修改城市定制信息
func (this *RegionProvinceService) UpdateRegionProvinceCustom(ctx context.Context, req *pb.UpdateRegionProvinceCustomRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = regions.SharedRegionProvinceDAO.UpdateProvinceCustom(tx, req.RegionProvinceId, req.CustomName, req.CustomCodes)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
