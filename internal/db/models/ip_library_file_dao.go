package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/iplibrary"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"io"
	"os"
	"strings"
	"time"
)

const (
	IPLibraryFileStateEnabled  = 1 // 已启用
	IPLibraryFileStateDisabled = 0 // 已禁用
)

type IPLibraryFileDAO dbs.DAO

func NewIPLibraryFileDAO() *IPLibraryFileDAO {
	return dbs.NewDAO(&IPLibraryFileDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeIPLibraryFiles",
			Model:  new(IPLibraryFile),
			PkName: "id",
		},
	}).(*IPLibraryFileDAO)
}

var SharedIPLibraryFileDAO *IPLibraryFileDAO

func init() {
	dbs.OnReady(func() {
		SharedIPLibraryFileDAO = NewIPLibraryFileDAO()
	})
}

// EnableIPLibraryFile 启用条目
func (this *IPLibraryFileDAO) EnableIPLibraryFile(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPLibraryFileStateEnabled).
		Update()
	return err
}

// DisableIPLibraryFile 禁用条目
func (this *IPLibraryFileDAO) DisableIPLibraryFile(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPLibraryFileStateDisabled).
		Update()
	return err
}

// FindEnabledIPLibraryFile 查找启用中的条目
func (this *IPLibraryFileDAO) FindEnabledIPLibraryFile(tx *dbs.Tx, id int64) (*IPLibraryFile, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(IPLibraryFileStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*IPLibraryFile), err
}

// CreateLibraryFile 创建文件
func (this *IPLibraryFileDAO) CreateLibraryFile(tx *dbs.Tx, name string, template string, emptyValues []string, password string, fileId int64, countries []string, provinces [][2]string, cities [][3]string, towns [][4]string, providers []string) (int64, error) {
	var op = NewIPLibraryFileOperator()
	op.Name = name
	op.Template = template

	if emptyValues == nil {
		emptyValues = []string{}
	}
	emptyValuesJSON, err := json.Marshal(emptyValues)
	if err != nil {
		return 0, err
	}
	op.EmptyValues = emptyValuesJSON

	op.Password = password

	op.FileId = fileId

	if countries == nil {
		countries = []string{}
	}
	countriesJSON, err := json.Marshal(countries)
	if err != nil {
		return 0, err
	}
	op.Countries = countriesJSON

	if provinces == nil {
		provinces = [][2]string{}
	}
	provincesJSON, err := json.Marshal(provinces)
	if err != nil {
		return 0, err
	}
	op.Provinces = provincesJSON

	if cities == nil {
		cities = [][3]string{}
	}
	citiesJSON, err := json.Marshal(cities)
	if err != nil {
		return 0, err
	}
	op.Cities = citiesJSON

	if towns == nil {
		towns = [][4]string{}
	}
	townsJSON, err := json.Marshal(towns)
	if err != nil {
		return 0, err
	}
	op.Towns = townsJSON

	if providers == nil {
		providers = []string{}
	}
	providersJSON, err := json.Marshal(providers)
	if err != nil {
		return 0, err
	}
	op.Providers = providersJSON

	op.IsFinished = false
	op.State = IPLibraryFileStateEnabled
	return this.SaveInt64(tx, op)
}

// FindAllFinishedLibraryFiles 查找所有已完成的文件
func (this *IPLibraryFileDAO) FindAllFinishedLibraryFiles(tx *dbs.Tx) (result []*IPLibraryFile, err error) {
	_, err = this.Query(tx).
		State(IPLibraryFileStateEnabled).
		Result("id", "fileId", "createdAt", "generatedFileId", "generatedAt", "name"). // 这里不需要其他信息
		Attr("isFinished", true).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllUnfinishedLibraryFiles 查找所有未完成的文件
func (this *IPLibraryFileDAO) FindAllUnfinishedLibraryFiles(tx *dbs.Tx) (result []*IPLibraryFile, err error) {
	_, err = this.Query(tx).
		State(IPLibraryFileStateEnabled).
		Result("id", "fileId", "createdAt"). // 这里不需要其他信息
		Attr("isFinished", false).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// UpdateLibraryFileIsFinished 设置文件为已完成
func (this *IPLibraryFileDAO) UpdateLibraryFileIsFinished(tx *dbs.Tx, fileId int64) error {
	return this.Query(tx).
		Pk(fileId).
		Set("isFinished", true).
		UpdateQuickly()
}

// FindLibraryFileCountries 获取IP库中的国家/地区
func (this *IPLibraryFileDAO) FindLibraryFileCountries(tx *dbs.Tx, fileId int64) ([]string, error) {
	countriesJSON, err := this.Query(tx).
		Result("countries").
		Pk(fileId).
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	if IsNull(countriesJSON) {
		return nil, nil
	}

	var result = []string{}
	err = json.Unmarshal(countriesJSON, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindLibraryFileProvinces 获取IP库中的省份
func (this *IPLibraryFileDAO) FindLibraryFileProvinces(tx *dbs.Tx, fileId int64) ([][2]string, error) {
	provincesJSON, err := this.Query(tx).
		Result("provinces").
		Pk(fileId).
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	if IsNull(provincesJSON) {
		return nil, nil
	}

	var result = [][2]string{}
	err = json.Unmarshal(provincesJSON, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindLibraryFileCities 获取IP库中的城市
func (this *IPLibraryFileDAO) FindLibraryFileCities(tx *dbs.Tx, fileId int64) ([][3]string, error) {
	citiesJSON, err := this.Query(tx).
		Result("cities").
		Pk(fileId).
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	if IsNull(citiesJSON) {
		return nil, nil
	}

	var result = [][3]string{}
	err = json.Unmarshal(citiesJSON, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindLibraryFileTowns 获取IP库中的区县
func (this *IPLibraryFileDAO) FindLibraryFileTowns(tx *dbs.Tx, fileId int64) ([][4]string, error) {
	townsJSON, err := this.Query(tx).
		Result("towns").
		Pk(fileId).
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	if IsNull(townsJSON) {
		return nil, nil
	}

	var result = [][4]string{}
	err = json.Unmarshal(townsJSON, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindLibraryFileProviders 获取IP库中的ISP运营商
func (this *IPLibraryFileDAO) FindLibraryFileProviders(tx *dbs.Tx, fileId int64) ([]string, error) {
	providersJSON, err := this.Query(tx).
		Result("providers").
		Pk(fileId).
		FindJSONCol()
	if err != nil {
		return nil, err
	}

	if IsNull(providersJSON) {
		return nil, nil
	}

	var result = []string{}
	err = json.Unmarshal(providersJSON, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (this *IPLibraryFileDAO) GenerateIPLibrary(tx *dbs.Tx, libraryFileId int64) error {
	one, err := this.Query(tx).Pk(libraryFileId).Find()
	if err != nil {
		return err
	}
	if one == nil {
		return errors.New("the library file not found")
	}

	var libraryFile = one.(*IPLibraryFile)
	template, err := iplibrary.NewTemplate(libraryFile.Template)
	if err != nil {
		return fmt.Errorf("create template from '%s' failed: %w", libraryFile.Template, err)
	}

	var fileId = int64(libraryFile.FileId)
	if fileId == 0 {
		return errors.New("the library file has not been uploaded yet")
	}

	var dir = Tea.Root + "/data"
	stat, err := os.Stat(dir)

	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0777)
			if err != nil {
				return fmt.Errorf("can not open dir '%s' to write: %w", dir, err)
			}
		} else {
			return fmt.Errorf("can not open dir '%s' to write: %w", dir, err)
		}
	} else if !stat.IsDir() {
		_ = os.Remove(dir)

		err = os.Mkdir(dir, 0777)
		if err != nil {
			return fmt.Errorf("can not open dir '%s' to write: %w", dir, err)
		}
	}

	// TODO 删除以往生成的文件，但要考虑到文件正在被别的任务所使用

	// 国家
	dbCountries, err := regions.SharedRegionCountryDAO.FindAllCountries(tx)
	if err != nil {
		return err
	}

	var countries = []*iplibrary.Country{}
	for _, country := range dbCountries {
		countries = append(countries, &iplibrary.Country{
			Id:    types.Uint16(country.ValueId),
			Name:  country.DisplayName(),
			Codes: country.AllCodes(),
		})
	}

	// 省份
	dbProvinces, err := regions.SharedRegionProvinceDAO.FindAllEnabledProvinces(tx)
	if err != nil {
		return err
	}

	var provinces = []*iplibrary.Province{}
	for _, province := range dbProvinces {
		provinces = append(provinces, &iplibrary.Province{
			Id:    types.Uint16(province.ValueId),
			Name:  province.DisplayName(),
			Codes: province.AllCodes(),
		})
	}

	// 城市
	dbCities, err := regions.SharedRegionCityDAO.FindAllEnabledCities(tx)
	if err != nil {
		return err
	}

	var cities = []*iplibrary.City{}
	for _, city := range dbCities {
		cities = append(cities, &iplibrary.City{
			Id:    city.ValueId,
			Name:  city.DisplayName(),
			Codes: city.AllCodes(),
		})
	}

	// 区县
	dbTowns, err := regions.SharedRegionTownDAO.FindAllRegionTowns(tx)
	if err != nil {
		return err
	}

	var towns = []*iplibrary.Town{}
	for _, town := range dbTowns {
		towns = append(towns, &iplibrary.Town{
			Id:    town.ValueId,
			Name:  town.DisplayName(),
			Codes: town.AllCodes(),
		})
	}

	// ISP运营商
	dbProviders, err := regions.SharedRegionProviderDAO.FindAllEnabledProviders(tx)
	if err != nil {
		return err
	}

	var providers = []*iplibrary.Provider{}
	for _, provider := range dbProviders {
		providers = append(providers, &iplibrary.Provider{
			Id:    types.Uint16(provider.ValueId),
			Name:  provider.DisplayName(),
			Codes: provider.AllCodes(),
		})
	}

	var libraryCode = utils.Sha1RandomString() // 每次都生成新的code
	var filePath = dir + "/" + this.composeFilename(libraryFileId, libraryCode)
	var meta = &iplibrary.Meta{
		Author:    "", // 将来用户可以自行填写
		CreatedAt: time.Now().Unix(),
		Countries: countries,
		Provinces: provinces,
		Cities:    cities,
		Towns:     towns,
		Providers: providers,
	}
	writer, err := iplibrary.NewFileWriter(filePath, meta, libraryFile.Password)
	if err != nil {
		return err
	}

	defer func() {
		_ = writer.Close()
		_ = os.Remove(filePath)
	}()

	err = writer.WriteMeta()
	if err != nil {
		return fmt.Errorf("write meta failed: %w", err)
	}

	chunkIds, err := SharedFileChunkDAO.FindAllFileChunkIds(tx, fileId)
	if err != nil {
		return err
	}

	// countries etc ...
	var countryMap = map[string]int64{} // countryName => countryId
	for _, country := range dbCountries {
		for _, code := range country.AllCodes() {
			countryMap[code] = int64(country.ValueId)
		}
	}

	var provinceMap = map[string]int64{} // countryId_provinceName => provinceId
	for _, province := range dbProvinces {
		for _, code := range province.AllCodes() {
			provinceMap[types.String(province.CountryId)+"_"+code] = int64(province.ValueId)

			for _, suffix := range regions.RegionProvinceSuffixes {
				if strings.HasSuffix(code, suffix) {
					provinceMap[types.String(province.CountryId)+"_"+strings.TrimSuffix(code, suffix)] = int64(province.ValueId)
				} else {
					provinceMap[types.String(province.CountryId)+"_"+(code+suffix)] = int64(province.ValueId)
				}
			}
		}
	}

	var cityMap = map[string]int64{} // provinceId_cityName => cityId
	for _, city := range dbCities {
		for _, code := range city.AllCodes() {
			cityMap[types.String(city.ProvinceId)+"_"+code] = int64(city.ValueId)
		}
	}

	var townMap = map[string]int64{} // cityId_townName => townId
	for _, town := range dbTowns {
		for _, code := range town.AllCodes() {
			townMap[types.String(town.CityId)+"_"+code] = int64(town.ValueId)
		}
	}

	var providerMap = map[string]int64{} // providerName => providerId
	for _, provider := range dbProviders {
		for _, code := range provider.AllCodes() {
			providerMap[code] = int64(provider.ValueId)
		}
	}

	dataParser, err := iplibrary.NewParser(&iplibrary.ParserConfig{
		Template:    template,
		EmptyValues: libraryFile.DecodeEmptyValues(),
		Iterator: func(values map[string]string) error {
			var ipFrom = values["ipFrom"]
			var ipTo = values["ipTo"]

			var countryName = values["country"]
			var provinceName = values["province"]
			var cityName = values["city"]
			var townName = values["town"]
			var providerName = values["provider"]

			var countryId = countryMap[countryName]
			var provinceId int64
			var cityId int64 = 0
			var townId int64 = 0
			var providerId = providerMap[providerName]

			if countryId > 0 {
				provinceId = provinceMap[types.String(countryId)+"_"+provinceName]
				if provinceId > 0 {
					cityId = cityMap[types.String(provinceId)+"_"+cityName]
					if cityId > 0 {
						townId = townMap[types.String(cityId)+"_"+townName]
					}
				}
			}

			err = writer.Write(ipFrom, ipTo, countryId, provinceId, cityId, townId, providerId)
			if err != nil {
				return fmt.Errorf("write failed: %w", err)
			}

			return nil
		},
	})
	if err != nil {
		return err
	}

	for _, chunkId := range chunkIds {
		chunk, err := SharedFileChunkDAO.FindFileChunk(tx, chunkId)
		if err != nil {
			return err
		}
		if chunk == nil {
			return errors.New("invalid chunk file, please upload again")
		}
		dataParser.Write(chunk.Data)
		err = dataParser.Parse()
		if err != nil {
			return err
		}
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	// 将生成的内容写入到文件
	stat, err = os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat generated file failed: %w", err)
	}
	generatedFileId, err := SharedFileDAO.CreateFile(tx, 0, 0, "ipLibraryFile", "", libraryCode+".db", stat.Size(), "", false)
	if err != nil {
		return err
	}

	fp, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open generated file failed: %w", err)
	}
	var buf = make([]byte, 256*1024)
	for {
		n, err := fp.Read(buf)
		if n > 0 {
			_, err = SharedFileChunkDAO.CreateFileChunk(tx, generatedFileId, buf[:n])
			if err != nil {
				return err
			}
		}
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
	}
	err = SharedFileDAO.UpdateFileIsFinished(tx, generatedFileId)
	if err != nil {
		return err
	}

	// 设置code
	err = this.Query(tx).
		Pk(libraryFileId).
		Set("code", libraryCode).
		Set("isFinished", true).
		Set("generatedFileId", generatedFileId).
		Set("generatedAt", time.Now().Unix()).
		UpdateQuickly()
	if err != nil {
		return err
	}

	// 添加制品
	_, err = SharedIPLibraryArtifactDAO.CreateArtifact(tx, libraryFile.Name, generatedFileId, libraryFileId, meta)
	if err != nil {
		return err
	}

	return nil
}

// 组合IP库文件名
func (this *IPLibraryFileDAO) composeFilename(libraryFileId int64, code string) string {
	return "ip-library-" + types.String(libraryFileId) + "-" + code + ".db"
}
