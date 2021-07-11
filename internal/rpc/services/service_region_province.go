package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// RegionProvinceService 省份相关服务
type RegionProvinceService struct {
	BaseService
}

// FindAllEnabledRegionProvincesWithCountryId 查找所有省份
func (this *RegionProvinceService) FindAllEnabledRegionProvincesWithCountryId(ctx context.Context, req *pb.FindAllEnabledRegionProvincesWithCountryIdRequest) (*pb.FindAllEnabledRegionProvincesWithCountryIdResponse, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	provinces, err := regions.SharedRegionProvinceDAO.FindAllEnabledProvincesWithCountryId(tx, req.CountryId)
	if err != nil {
		return nil, err
	}
	result := []*pb.RegionProvince{}
	for _, province := range provinces {
		result = append(result, &pb.RegionProvince{
			Id:    int64(province.Id),
			Name:  province.Name,
			Codes: province.DecodeCodes(),
		})
	}

	return &pb.FindAllEnabledRegionProvincesWithCountryIdResponse{
		Provinces: result,
	}, nil
}

// FindEnabledRegionProvince 查找单个省份信息
func (this *RegionProvinceService) FindEnabledRegionProvince(ctx context.Context, req *pb.FindEnabledRegionProvinceRequest) (*pb.FindEnabledRegionProvinceResponse, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	province, err := regions.SharedRegionProvinceDAO.FindEnabledRegionProvince(tx, req.ProvinceId)
	if err != nil {
		return nil, err
	}
	if province == nil {
		return &pb.FindEnabledRegionProvinceResponse{Province: nil}, nil
	}

	return &pb.FindEnabledRegionProvinceResponse{
		Province: &pb.RegionProvince{
			Id:    int64(province.Id),
			Name:  province.Name,
			Codes: province.DecodeCodes(),
		},
	}, nil
}
