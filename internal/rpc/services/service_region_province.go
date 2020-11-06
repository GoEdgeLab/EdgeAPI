package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 省份相关服务
type RegionProvinceService struct {
}

// 查找所有省份
func (this *RegionProvinceService) FindAllEnabledRegionProvincesWithCountryId(ctx context.Context, req *pb.FindAllEnabledRegionProvincesWithCountryIdRequest) (*pb.FindAllEnabledRegionProvincesWithCountryIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	provinces, err := models.SharedRegionProvinceDAO.FindAllEnabledProvincesWithCountryId(req.CountryId)
	if err != nil {
		return nil, err
	}
	result := []*pb.RegionProvince{}
	for _, province := range provinces {
		result = append(result, &pb.RegionProvince{
			Id:   int64(province.Id),
			Name: province.Name,
		})
	}

	return &pb.FindAllEnabledRegionProvincesWithCountryIdResponse{
		Provinces: result,
	}, nil
}
