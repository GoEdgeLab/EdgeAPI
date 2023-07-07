package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// RegionCountryService 国家相关服务
type RegionCountryService struct {
	BaseService
}

// FindAllEnabledRegionCountries 查找所有的国家列表
// Deprecated
func (this *RegionCountryService) FindAllEnabledRegionCountries(ctx context.Context, req *pb.FindAllEnabledRegionCountriesRequest) (*pb.FindAllEnabledRegionCountriesResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	countries, err := regions.SharedRegionCountryDAO.FindAllEnabledCountriesOrderByPinyin(tx)
	if err != nil {
		return nil, err
	}

	result := []*pb.RegionCountry{}
	for _, country := range countries {
		pinyinStrings := []string{}
		err = json.Unmarshal(country.Pinyin, &pinyinStrings)
		if err != nil {
			return nil, err
		}
		if len(pinyinStrings) == 0 {
			continue
		}

		result = append(result, &pb.RegionCountry{
			Id:          int64(country.ValueId),
			Name:        country.Name,
			Codes:       country.DecodeCodes(),
			Pinyin:      pinyinStrings,
			CustomName:  country.CustomName,
			CustomCodes: country.DecodeCustomCodes(),
			DisplayName: country.DisplayName(),
			IsCommon:    country.IsCommon,
		})
	}
	return &pb.FindAllEnabledRegionCountriesResponse{
		RegionCountries: result,
	}, nil
}

// FindEnabledRegionCountry 查找单个国家信息
// Deprecated
func (this *RegionCountryService) FindEnabledRegionCountry(ctx context.Context, req *pb.FindEnabledRegionCountryRequest) (*pb.FindEnabledRegionCountryResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	country, err := regions.SharedRegionCountryDAO.FindEnabledRegionCountry(tx, req.RegionCountryId)
	if err != nil {
		return nil, err
	}
	if country == nil {
		return &pb.FindEnabledRegionCountryResponse{RegionCountry: nil}, nil
	}

	return &pb.FindEnabledRegionCountryResponse{RegionCountry: &pb.RegionCountry{
		Id:          int64(country.ValueId),
		Name:        country.Name,
		Codes:       country.DecodeCodes(),
		CustomName:  country.CustomName,
		CustomCodes: country.DecodeCustomCodes(),
		DisplayName: country.DisplayName(),
	}}, nil
}

// FindAllRegionCountries 查找所有的国家列表
func (this *RegionCountryService) FindAllRegionCountries(ctx context.Context, req *pb.FindAllRegionCountriesRequest) (*pb.FindAllRegionCountriesResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	countries, err := regions.SharedRegionCountryDAO.FindAllEnabledCountriesOrderByPinyin(tx)
	if err != nil {
		return nil, err
	}

	var result = []*pb.RegionCountry{}
	for _, country := range countries {
		pinyinStrings := []string{}
		err = json.Unmarshal(country.Pinyin, &pinyinStrings)
		if err != nil {
			return nil, err
		}
		if len(pinyinStrings) == 0 {
			continue
		}

		result = append(result, &pb.RegionCountry{
			Id:          int64(country.ValueId),
			Name:        country.Name,
			Codes:       country.DecodeCodes(),
			Pinyin:      pinyinStrings,
			CustomName:  country.CustomName,
			CustomCodes: country.DecodeCustomCodes(),
			DisplayName: country.DisplayName(),
			IsCommon:    country.IsCommon,
		})
	}
	return &pb.FindAllRegionCountriesResponse{
		RegionCountries: result,
	}, nil
}

// FindRegionCountry 查找单个国家信息
func (this *RegionCountryService) FindRegionCountry(ctx context.Context, req *pb.FindRegionCountryRequest) (*pb.FindRegionCountryResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	country, err := regions.SharedRegionCountryDAO.FindEnabledRegionCountry(tx, req.RegionCountryId)
	if err != nil {
		return nil, err
	}
	if country == nil {
		return &pb.FindRegionCountryResponse{RegionCountry: nil}, nil
	}

	return &pb.FindRegionCountryResponse{RegionCountry: &pb.RegionCountry{
		Id:          int64(country.ValueId),
		Name:        country.Name,
		Codes:       country.DecodeCodes(),
		CustomName:  country.CustomName,
		CustomCodes: country.DecodeCustomCodes(),
		DisplayName: country.DisplayName(),
		IsCommon:    country.IsCommon,
	}}, nil
}

// UpdateRegionCountryCustom 修改城市定制信息
func (this *RegionCountryService) UpdateRegionCountryCustom(ctx context.Context, req *pb.UpdateRegionCountryCustomRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = regions.SharedRegionCountryDAO.UpdateCountryCustom(tx, req.RegionCountryId, req.CustomName, req.CustomCodes)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
