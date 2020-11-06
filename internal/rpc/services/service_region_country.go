package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 国家相关服务
type RegionCountryService struct {
}

// 查找所有的国家列表
func (this *RegionCountryService) FindAllEnabledRegionCountries(ctx context.Context, req *pb.FindAllEnabledRegionCountriesRequest) (*pb.FindAllEnabledRegionCountriesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	countries, err := models.SharedRegionCountryDAO.FindAllEnabledCountriesOrderByPinyin()
	if err != nil {
		return nil, err
	}

	result := []*pb.RegionCountry{}
	for _, country := range countries {
		pinyinStrings := []string{}
		err = json.Unmarshal([]byte(country.Pinyin), &pinyinStrings)
		if err != nil {
			return nil, err
		}
		if len(pinyinStrings) == 0 {
			continue
		}

		result = append(result, &pb.RegionCountry{
			Id:     int64(country.Id),
			Name:   country.Name,
			Pinyin: pinyinStrings,
		})
	}
	return &pb.FindAllEnabledRegionCountriesResponse{
		Countries: result,
	}, nil
}
