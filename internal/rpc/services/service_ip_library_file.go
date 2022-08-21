// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
)

// IPLibraryFileService IP库文件管理
type IPLibraryFileService struct {
	BaseService
}

// FindAllFinishedIPLibraryFiles 查找所有已完成的IP库文件
func (this *IPLibraryFileService) FindAllFinishedIPLibraryFiles(ctx context.Context, req *pb.FindAllFinishedIPLibraryFilesRequest) (*pb.FindAllFinishedIPLibraryFilesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	libraryFiles, err := models.SharedIPLibraryFileDAO.FindAllFinishedLibraryFiles(tx)
	if err != nil {
		return nil, err
	}
	var pbLibraryFiles = []*pb.IPLibraryFile{}
	for _, libraryFile := range libraryFiles {
		var pbCountryNames = libraryFile.DecodeCountries()
		var pbProviderNames = libraryFile.DecodeProviders()

		var pbProvinces = []*pb.IPLibraryFile_Province{}
		for _, province := range libraryFile.DecodeProvinces() {
			pbProvinces = append(pbProvinces, &pb.IPLibraryFile_Province{
				CountryName:  province[0],
				ProvinceName: province[1],
			})
		}

		var pbCities = []*pb.IPLibraryFile_City{}
		for _, city := range libraryFile.DecodeCities() {
			pbCities = append(pbCities, &pb.IPLibraryFile_City{
				CountryName:  city[0],
				ProvinceName: city[1],
				CityName:     city[2],
			})
		}

		var pbTowns = []*pb.IPLibraryFile_Town{}
		for _, town := range libraryFile.DecodeTowns() {
			pbTowns = append(pbTowns, &pb.IPLibraryFile_Town{
				CountryName:  town[0],
				ProvinceName: town[1],
				CityName:     town[2],
				TownName:     town[3],
			})
		}

		pbLibraryFiles = append(pbLibraryFiles, &pb.IPLibraryFile{
			Id:              int64(libraryFile.Id),
			Name:            libraryFile.Name,
			FileId:          int64(libraryFile.FileId),
			IsFinished:      libraryFile.IsFinished,
			CreatedAt:       int64(libraryFile.CreatedAt),
			GeneratedFileId: int64(libraryFile.GeneratedFileId),
			GeneratedAt:     int64(libraryFile.GeneratedAt),
			CountryNames:    pbCountryNames,
			Provinces:       pbProvinces,
			Cities:          pbCities,
			Towns:           pbTowns,
			ProviderNames:   pbProviderNames,
		})
	}

	return &pb.FindAllFinishedIPLibraryFilesResponse{
		IpLibraryFiles: pbLibraryFiles,
	}, nil
}

// FindAllUnfinishedIPLibraryFiles 查找所有未完成的IP库文件
func (this *IPLibraryFileService) FindAllUnfinishedIPLibraryFiles(ctx context.Context, req *pb.FindAllUnfinishedIPLibraryFilesRequest) (*pb.FindAllUnfinishedIPLibraryFilesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	libraryFiles, err := models.SharedIPLibraryFileDAO.FindAllUnfinishedLibraryFiles(tx)
	if err != nil {
		return nil, err
	}
	var pbLibraryFiles = []*pb.IPLibraryFile{}
	for _, libraryFile := range libraryFiles {
		var pbCountryNames = libraryFile.DecodeCountries()
		var pbProviderNames = libraryFile.DecodeProviders()

		var pbProvinces = []*pb.IPLibraryFile_Province{}
		for _, province := range libraryFile.DecodeProvinces() {
			pbProvinces = append(pbProvinces, &pb.IPLibraryFile_Province{
				CountryName:  province[0],
				ProvinceName: province[1],
			})
		}

		var pbCities = []*pb.IPLibraryFile_City{}
		for _, city := range libraryFile.DecodeCities() {
			pbCities = append(pbCities, &pb.IPLibraryFile_City{
				CountryName:  city[0],
				ProvinceName: city[1],
				CityName:     city[2],
			})
		}

		var pbTowns = []*pb.IPLibraryFile_Town{}
		for _, town := range libraryFile.DecodeTowns() {
			pbTowns = append(pbTowns, &pb.IPLibraryFile_Town{
				CountryName:  town[0],
				ProvinceName: town[1],
				CityName:     town[2],
				TownName:     town[3],
			})
		}

		pbLibraryFiles = append(pbLibraryFiles, &pb.IPLibraryFile{
			Id:            int64(libraryFile.Id),
			Name:          libraryFile.Name,
			FileId:        int64(libraryFile.FileId),
			IsFinished:    libraryFile.IsFinished,
			CreatedAt:     int64(libraryFile.CreatedAt),
			CountryNames:  pbCountryNames,
			Provinces:     pbProvinces,
			Cities:        pbCities,
			Towns:         pbTowns,
			ProviderNames: pbProviderNames,
		})
	}

	return &pb.FindAllUnfinishedIPLibraryFilesResponse{
		IpLibraryFiles: pbLibraryFiles,
	}, nil
}

// FindIPLibraryFile 查找单个IP库文件
func (this *IPLibraryFileService) FindIPLibraryFile(ctx context.Context, req *pb.FindIPLibraryFileRequest) (*pb.FindIPLibraryFileResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	libraryFile, err := models.SharedIPLibraryFileDAO.FindEnabledIPLibraryFile(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}
	if libraryFile == nil {
		return &pb.FindIPLibraryFileResponse{
			IpLibraryFile: nil,
		}, nil
	}

	var pbCountryNames = libraryFile.DecodeCountries()
	var pbProviderNames = libraryFile.DecodeProviders()

	var pbProvinces = []*pb.IPLibraryFile_Province{}
	for _, province := range libraryFile.DecodeProvinces() {
		pbProvinces = append(pbProvinces, &pb.IPLibraryFile_Province{
			CountryName:  province[0],
			ProvinceName: province[1],
		})
	}

	var pbCities = []*pb.IPLibraryFile_City{}
	for _, city := range libraryFile.DecodeCities() {
		pbCities = append(pbCities, &pb.IPLibraryFile_City{
			CountryName:  city[0],
			ProvinceName: city[1],
			CityName:     city[2],
		})
	}

	var pbTowns = []*pb.IPLibraryFile_Town{}
	for _, town := range libraryFile.DecodeTowns() {
		pbTowns = append(pbTowns, &pb.IPLibraryFile_Town{
			CountryName:  town[0],
			ProvinceName: town[1],
			CityName:     town[2],
			TownName:     town[3],
		})
	}

	return &pb.FindIPLibraryFileResponse{
		IpLibraryFile: &pb.IPLibraryFile{
			Id:              int64(libraryFile.Id),
			Name:            libraryFile.Name,
			Template:        libraryFile.Template,
			EmptyValues:     libraryFile.DecodeEmptyValues(),
			FileId:          int64(libraryFile.FileId),
			IsFinished:      libraryFile.IsFinished,
			CreatedAt:       int64(libraryFile.CreatedAt),
			GeneratedFileId: int64(libraryFile.GeneratedFileId),
			CountryNames:    pbCountryNames,
			Provinces:       pbProvinces,
			Cities:          pbCities,
			Towns:           pbTowns,
			ProviderNames:   pbProviderNames,
		},
	}, nil
}

// CreateIPLibraryFile 创建IP库文件
func (this *IPLibraryFileService) CreateIPLibraryFile(ctx context.Context, req *pb.CreateIPLibraryFileRequest) (*pb.CreateIPLibraryFileResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var countries = []string{}
	var provinces = [][2]string{}
	var cities = [][3]string{}
	var towns = [][4]string{}
	var providers = []string{}

	err = json.Unmarshal(req.CountriesJSON, &countries)
	if err != nil {
		return nil, errors.New("decode countries failed: " + err.Error())
	}

	err = json.Unmarshal(req.ProvincesJSON, &provinces)
	if err != nil {
		return nil, errors.New("decode provinces failed: " + err.Error())
	}

	err = json.Unmarshal(req.CitiesJSON, &cities)
	if err != nil {
		return nil, errors.New("decode cities failed: " + err.Error())
	}

	err = json.Unmarshal(req.TownsJSON, &towns)
	if err != nil {
		return nil, errors.New("decode towns failed: " + err.Error())
	}

	err = json.Unmarshal(req.ProvidersJSON, &providers)
	if err != nil {
		return nil, errors.New("decode providers failed: " + err.Error())
	}

	var tx = this.NullTx()
	libraryFileId, err := models.SharedIPLibraryFileDAO.CreateLibraryFile(tx, req.Name, req.Template, req.EmptyValues, req.FileId, countries, provinces, cities, towns, providers)
	if err != nil {
		return nil, err
	}
	return &pb.CreateIPLibraryFileResponse{
		IpLibraryFileId: libraryFileId,
	}, nil
}

// CheckCountriesWithIPLibraryFileId 检查国家/地区
func (this *IPLibraryFileService) CheckCountriesWithIPLibraryFileId(ctx context.Context, req *pb.CheckCountriesWithIPLibraryFileIdRequest) (*pb.CheckCountriesWithIPLibraryFileIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	allCountries, err := regions.SharedRegionCountryDAO.FindAllCountries(tx)
	if err != nil {
		return nil, err
	}

	countryNames, err := models.SharedIPLibraryFileDAO.FindLibraryFileCountries(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}
	var pbMissingCountries = []*pb.CheckCountriesWithIPLibraryFileIdResponse_MissingCountry{}
	for _, countryName := range countryNames {
		if len(countryName) == 0 {
			continue
		}

		// 检查是否存在
		countryId, err := regions.SharedRegionCountryDAO.FindCountryIdWithName(tx, countryName)
		if err != nil {
			return nil, err
		}
		if countryId > 0 {
			continue
		}

		var pbMissingCountry = &pb.CheckCountriesWithIPLibraryFileIdResponse_MissingCountry{
			CountryName:      countryName,
			SimilarCountries: nil,
		}

		// 查找相似
		var similarCountries = regions.SharedRegionCountryDAO.FindSimilarCountries(allCountries, countryName, 5)
		if err != nil {
			return nil, err
		}
		for _, similarCountry := range similarCountries {
			pbMissingCountry.SimilarCountries = append(pbMissingCountry.SimilarCountries, &pb.RegionCountry{
				Id:          int64(similarCountry.Id),
				Name:        similarCountry.Name,
				DisplayName: similarCountry.DisplayName(),
			})
		}

		pbMissingCountries = append(pbMissingCountries, pbMissingCountry)
	}

	return &pb.CheckCountriesWithIPLibraryFileIdResponse{
		MissingCountries: pbMissingCountries,
	}, nil
}

// CheckProvincesWithIPLibraryFileId 检查省份/州
func (this *IPLibraryFileService) CheckProvincesWithIPLibraryFileId(ctx context.Context, req *pb.CheckProvincesWithIPLibraryFileIdRequest) (*pb.CheckProvincesWithIPLibraryFileIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	provinces, err := models.SharedIPLibraryFileDAO.FindLibraryFileProvinces(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}
	var countryMap = map[string]int64{}            // countryName => countryId
	var provinceNamesMap = map[int64][][2]string{} // countryId => [][2]{countryName, provinceName}
	var countryIds = []int64{}
	for _, province := range provinces {
		var countryName = province[0]
		var provinceName = province[1]

		countryId, ok := countryMap[countryName]
		if ok {
			provinceNamesMap[countryId] = append(provinceNamesMap[countryId], [2]string{countryName, provinceName})
			continue
		}

		countryId, err := regions.SharedRegionCountryDAO.FindCountryIdWithName(tx, countryName)
		if err != nil {
			return nil, err
		}

		countryMap[countryName] = countryId

		provinceNamesMap[countryId] = append(provinceNamesMap[countryId], [2]string{countryName, provinceName})

		if countryId > 0 && !lists.ContainsInt64(countryIds, countryId) {
			countryIds = append(countryIds, countryId)
		}
	}

	var pbMissingProvinces = []*pb.CheckProvincesWithIPLibraryFileIdResponse_MissingProvince{}
	for _, countryId := range countryIds {
		allProvinces, err := regions.SharedRegionProvinceDAO.FindAllEnabledProvincesWithCountryId(tx, countryId)
		if err != nil {
			return nil, err
		}

		for _, province := range provinceNamesMap[countryId] {
			var countryName = province[0]
			var provinceName = province[1]
			provinceId, err := regions.SharedRegionProvinceDAO.FindProvinceIdWithName(tx, countryId, provinceName)
			if err != nil {
				return nil, err
			}
			if provinceId > 0 {
				continue
			}

			var similarProvinces = regions.SharedRegionProvinceDAO.FindSimilarProvinces(allProvinces, provinceName, 5)
			if err != nil {
				return nil, err
			}
			var pbMissingProvince = &pb.CheckProvincesWithIPLibraryFileIdResponse_MissingProvince{}
			pbMissingProvince.CountryName = countryName
			pbMissingProvince.ProvinceName = provinceName

			for _, similarProvince := range similarProvinces {
				pbMissingProvince.SimilarProvinces = append(pbMissingProvince.SimilarProvinces, &pb.RegionProvince{
					Id:          int64(similarProvince.Id),
					Name:        similarProvince.Name,
					DisplayName: similarProvince.DisplayName(),
				})
			}
			pbMissingProvinces = append(pbMissingProvinces, pbMissingProvince)
		}
	}

	return &pb.CheckProvincesWithIPLibraryFileIdResponse{MissingProvinces: pbMissingProvinces}, nil
}

// CheckCitiesWithIPLibraryFileId 检查城市/市
func (this *IPLibraryFileService) CheckCitiesWithIPLibraryFileId(ctx context.Context, req *pb.CheckCitiesWithIPLibraryFileIdRequest) (*pb.CheckCitiesWithIPLibraryFileIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	cities, err := models.SharedIPLibraryFileDAO.FindLibraryFileCities(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}
	var countryMap = map[string]int64{}        // countryName => countryId
	var provinceMap = map[string]int64{}       // countryId_provinceName => provinceId
	var cityNamesMap = map[int64][][3]string{} // provinceId => [][3]{countryName, provinceName, cityName}
	var provinceIds = []int64{}
	for _, city := range cities {
		var countryName = city[0]
		var provinceName = city[1]
		var cityName = city[2]

		countryId, ok := countryMap[countryName]
		if !ok {
			countryId, err = regions.SharedRegionCountryDAO.FindCountryIdWithName(tx, countryName)
			if err != nil {
				return nil, err
			}
		}

		countryMap[countryName] = countryId

		var key = types.String(countryId) + "_" + provinceName
		provinceId, ok := provinceMap[key]
		if ok {
			cityNamesMap[provinceId] = append(cityNamesMap[provinceId], [3]string{countryName, provinceName, cityName})
		} else {
			provinceId, err := regions.SharedRegionProvinceDAO.FindProvinceIdWithName(tx, countryId, provinceName)
			if err != nil {
				return nil, err
			}
			provinceMap[key] = provinceId
			cityNamesMap[provinceId] = append(cityNamesMap[provinceId], [3]string{countryName, provinceName, cityName})
			if provinceId > 0 {
				provinceIds = append(provinceIds, provinceId)
			}
		}
	}

	var pbMissingCities = []*pb.CheckCitiesWithIPLibraryFileIdResponse_MissingCity{}
	for _, provinceId := range provinceIds {
		allCities, err := regions.SharedRegionCityDAO.FindAllEnabledCitiesWithProvinceId(tx, provinceId)
		if err != nil {
			return nil, err
		}

		for _, city := range cityNamesMap[provinceId] {
			var countryName = city[0]
			var provinceName = city[1]
			var cityName = city[2]
			cityId, err := regions.SharedRegionCityDAO.FindCityIdWithName(tx, provinceId, cityName)
			if err != nil {
				return nil, err
			}
			if cityId > 0 {
				continue
			}

			var similarCities = regions.SharedRegionCityDAO.FindSimilarCities(allCities, cityName, 5)
			if err != nil {
				return nil, err
			}
			var pbMissingCity = &pb.CheckCitiesWithIPLibraryFileIdResponse_MissingCity{}
			pbMissingCity.CountryName = countryName
			pbMissingCity.ProvinceName = provinceName
			pbMissingCity.CityName = cityName

			for _, similarCity := range similarCities {
				pbMissingCity.SimilarCities = append(pbMissingCity.SimilarCities, &pb.RegionCity{
					Id:          int64(similarCity.Id),
					Name:        similarCity.Name,
					DisplayName: similarCity.DisplayName(),
				})
			}
			pbMissingCities = append(pbMissingCities, pbMissingCity)
		}
	}

	return &pb.CheckCitiesWithIPLibraryFileIdResponse{MissingCities: pbMissingCities}, nil
}

// CheckTownsWithIPLibraryFileId 检查区县
func (this *IPLibraryFileService) CheckTownsWithIPLibraryFileId(ctx context.Context, req *pb.CheckTownsWithIPLibraryFileIdRequest) (*pb.CheckTownsWithIPLibraryFileIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	towns, err := models.SharedIPLibraryFileDAO.FindLibraryFileTowns(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}
	var countryMap = map[string]int64{}  // countryName => countryId
	var provinceMap = map[string]int64{} // countryId_provinceName => provinceId
	var cityMap = map[string]int64{}     // province_cityName => cityId

	var townNamesMap = map[int64][][4]string{} // cityId => [][4]{countryName, provinceName, cityName, townName}
	var cityIds = []int64{}
	for _, town := range towns {
		var countryName = town[0]
		var provinceName = town[1]
		var cityName = town[2]
		var townName = town[3]

		// country
		countryId, ok := countryMap[countryName]
		if !ok {
			countryId, err = regions.SharedRegionCountryDAO.FindCountryIdWithName(tx, countryName)
			if err != nil {
				return nil, err
			}
		}

		countryMap[countryName] = countryId

		// province
		var provinceKey = types.String(countryId) + "_" + provinceName
		provinceId, ok := provinceMap[provinceKey]
		if !ok {
			if countryId > 0 {
				provinceId, err = regions.SharedRegionProvinceDAO.FindProvinceIdWithName(tx, countryId, provinceName)
				if err != nil {
					return nil, err
				}
			}
			provinceMap[provinceKey] = provinceId
		}

		// city
		var cityKey = types.String(provinceId) + "_" + cityName
		cityId, ok := cityMap[cityKey]
		if !ok {
			if provinceId > 0 {
				cityId, err = regions.SharedRegionCityDAO.FindCityIdWithName(tx, provinceId, cityName)
				if err != nil {
					return nil, err
				}
			}
			cityMap[cityKey] = cityId
			if cityId > 0 {
				cityIds = append(cityIds, cityId)
			}
		}

		// town
		townNamesMap[cityId] = append(townNamesMap[cityId], [4]string{countryName, provinceName, cityName, townName})
	}

	var pbMissingTowns = []*pb.CheckTownsWithIPLibraryFileIdResponse_MissingTown{}
	for _, cityId := range cityIds {
		allTowns, err := regions.SharedRegionTownDAO.FindAllRegionTownsWithCityId(tx, cityId)
		if err != nil {
			return nil, err
		}

		for _, town := range townNamesMap[cityId] {
			var countryName = town[0]
			var provinceName = town[1]
			var cityName = town[2]
			var townName = town[3]

			townId, err := regions.SharedRegionTownDAO.FindTownIdWithName(tx, cityId, townName)
			if err != nil {
				return nil, err
			}
			if townId > 0 {
				// 已存在，则跳过
				continue
			}

			var similarTowns = regions.SharedRegionTownDAO.FindSimilarTowns(allTowns, townName, 5)
			if err != nil {
				return nil, err
			}
			var pbMissingTown = &pb.CheckTownsWithIPLibraryFileIdResponse_MissingTown{}
			pbMissingTown.CountryName = countryName
			pbMissingTown.ProvinceName = provinceName
			pbMissingTown.CityName = cityName
			pbMissingTown.TownName = townName

			for _, similarTown := range similarTowns {
				pbMissingTown.SimilarTowns = append(pbMissingTown.SimilarTowns, &pb.RegionTown{
					Id:          int64(similarTown.Id),
					Name:        similarTown.Name,
					DisplayName: similarTown.DisplayName(),
				})
			}
			pbMissingTowns = append(pbMissingTowns, pbMissingTown)
		}
	}

	return &pb.CheckTownsWithIPLibraryFileIdResponse{MissingTowns: pbMissingTowns}, nil
}

// CheckProvidersWithIPLibraryFileId 检查ISP运营商
func (this *IPLibraryFileService) CheckProvidersWithIPLibraryFileId(ctx context.Context, req *pb.CheckProvidersWithIPLibraryFileIdRequest) (*pb.CheckProvidersWithIPLibraryFileIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	allProviders, err := regions.SharedRegionProviderDAO.FindAllEnabledProviders(tx)
	if err != nil {
		return nil, err
	}

	providerNames, err := models.SharedIPLibraryFileDAO.FindLibraryFileProviders(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}
	var pbMissingProviders = []*pb.CheckProvidersWithIPLibraryFileIdResponse_MissingProvider{}
	for _, providerName := range providerNames {
		if len(providerName) == 0 {
			continue
		}

		// 检查是否存在
		providerId, err := regions.SharedRegionProviderDAO.FindProviderIdWithName(tx, providerName)
		if err != nil {
			return nil, err
		}
		if providerId > 0 {
			continue
		}

		var pbMissingProvider = &pb.CheckProvidersWithIPLibraryFileIdResponse_MissingProvider{
			ProviderName:     providerName,
			SimilarProviders: nil,
		}

		// 查找相似
		var similarProviders = regions.SharedRegionProviderDAO.FindSimilarProviders(allProviders, providerName, 5)
		if err != nil {
			return nil, err
		}
		for _, similarProvider := range similarProviders {
			pbMissingProvider.SimilarProviders = append(pbMissingProvider.SimilarProviders, &pb.RegionProvider{
				Id:          int64(similarProvider.Id),
				Name:        similarProvider.Name,
				DisplayName: similarProvider.DisplayName(),
			})
		}

		pbMissingProviders = append(pbMissingProviders, pbMissingProvider)
	}

	return &pb.CheckProvidersWithIPLibraryFileIdResponse{
		MissingProviders: pbMissingProviders,
	}, nil
}

// GenerateIPLibraryFile 生成IP库文件
func (this *IPLibraryFileService) GenerateIPLibraryFile(ctx context.Context, req *pb.GenerateIPLibraryFileRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedIPLibraryFileDAO.GenerateIPLibrary(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateIPLibraryFileFinished 设置某个IP库为已完成
func (this *IPLibraryFileService) UpdateIPLibraryFileFinished(ctx context.Context, req *pb.UpdateIPLibraryFileFinishedRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedIPLibraryFileDAO.UpdateLibraryFileIsFinished(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteIPLibraryFile 删除IP库文件
func (this *IPLibraryFileService) DeleteIPLibraryFile(ctx context.Context, req *pb.DeleteIPLibraryFileRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedIPLibraryFileDAO.DisableIPLibraryFile(tx, req.IpLibraryFileId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
