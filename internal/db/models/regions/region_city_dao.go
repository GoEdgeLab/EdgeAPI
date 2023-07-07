package regions

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"sort"
	"strconv"
)

const (
	RegionCityStateEnabled  = 1 // 已启用
	RegionCityStateDisabled = 0 // 已禁用
)

type RegionCityDAO dbs.DAO

func NewRegionCityDAO() *RegionCityDAO {
	return dbs.NewDAO(&RegionCityDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionCities",
			Model:  new(RegionCity),
			PkName: "id",
		},
	}).(*RegionCityDAO)
}

var SharedRegionCityDAO *RegionCityDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionCityDAO = NewRegionCityDAO()
	})
}

// EnableRegionCity 启用条目
func (this *RegionCityDAO) EnableRegionCity(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Attr("valueId", id).
		Set("state", RegionCityStateEnabled).
		Update()
	return err
}

// DisableRegionCity 禁用条目
func (this *RegionCityDAO) DisableRegionCity(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Attr("valueId", id).
		Set("state", RegionCityStateDisabled).
		Update()
	return err
}

// FindEnabledRegionCity 查找启用中的条目
func (this *RegionCityDAO) FindEnabledRegionCity(tx *dbs.Tx, id int64) (*RegionCity, error) {
	result, err := this.Query(tx).
		Attr("valueId", id).
		Attr("state", RegionCityStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionCity), err
}

// FindRegionCityName 根据主键查找名称
func (this *RegionCityDAO) FindRegionCityName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Attr("valueId", id).
		Result("name").
		FindStringCol("")
}

// FindCityWithDataId 根据数据ID查找城市
func (this *RegionCityDAO) FindCityWithDataId(tx *dbs.Tx, dataId string) (int64, error) {
	return this.Query(tx).
		Attr("dataId", dataId).
		Result(RegionCityField_ValueId).
		FindInt64Col(0)
}

// CreateCity 创建城市
func (this *RegionCityDAO) CreateCity(tx *dbs.Tx, provinceId int64, name string, dataId string) (int64, error) {
	var op = NewRegionCityOperator()
	op.ProvinceId = provinceId
	op.Name = name
	op.DataId = dataId
	op.State = RegionCityStateEnabled

	var codes = []string{name}
	codesJSON, err := json.Marshal(codes)
	if err != nil {
		return 0, err
	}
	op.Codes = codesJSON
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	var cityId = types.Int64(op.Id)

	// value id
	err = this.Query(tx).
		Pk(cityId).
		Set(RegionCityField_ValueId, cityId).
		UpdateQuickly()
	if err != nil {
		return 0, err
	}

	return cityId, nil
}

// FindCityIdWithName 根据城市名查找城市ID
func (this *RegionCityDAO) FindCityIdWithName(tx *dbs.Tx, provinceId int64, cityName string) (int64, error) {
	return this.Query(tx).
		Attr("provinceId", provinceId).
		Where("(name=:cityName OR customName=:cityName OR JSON_CONTAINS(codes, :cityNameJSON) OR JSON_CONTAINS(customCodes, :cityNameJSON))").
		Param("cityName", cityName).
		Param("cityNameJSON", strconv.Quote(cityName)). // 查询的需要是个JSON字符串，所以这里加双引号
		Result(RegionCityField_ValueId).
		FindInt64Col(0)
}

// FindAllEnabledCities 获取所有城市信息
func (this *RegionCityDAO) FindAllEnabledCities(tx *dbs.Tx) (result []*RegionCity, err error) {
	_, err = this.Query(tx).
		State(RegionCityStateEnabled).
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledCitiesWithProvinceId 获取某个省份下的所有城市
func (this *RegionCityDAO) FindAllEnabledCitiesWithProvinceId(tx *dbs.Tx, provinceId int64) (result []*RegionCity, err error) {
	_, err = this.Query(tx).
		Attr("provinceId", provinceId).
		State(RegionCityStateEnabled).
		Slice(&result).
		FindAll()
	return
}

// UpdateCityCustom 自定义城市信息
func (this *RegionCityDAO) UpdateCityCustom(tx *dbs.Tx, cityId int64, customName string, customCodes []string) error {
	if customCodes == nil {
		customCodes = []string{}
	}
	customCodesJSON, err := json.Marshal(customCodes)
	if err != nil {
		return err
	}

	return this.Query(tx).
		Attr(RegionCityField_ValueId, cityId).
		Set("customName", customName).
		Set("customCodes", customCodesJSON).
		UpdateQuickly()
}

// FindSimilarCities 查找类似城市名
func (this *RegionCityDAO) FindSimilarCities(cities []*RegionCity, cityName string, size int) (result []*RegionCity) {
	if len(cities) == 0 {
		return
	}

	var similarResult = []maps.Map{}

	for _, city := range cities {
		var similarityList = []float32{}
		for _, code := range city.AllCodes() {
			var similarity = utils.Similar(cityName, code)
			if similarity > 0 {
				similarityList = append(similarityList, similarity)
			}
		}
		if len(similarityList) > 0 {
			similarResult = append(similarResult, maps.Map{
				"similarity": numberutils.Max(similarityList...),
				"city":       city,
			})
		}
	}

	sort.Slice(similarResult, func(i, j int) bool {
		return similarResult[i].GetFloat32("similarity") > similarResult[j].GetFloat32("similarity")
	})

	if len(similarResult) > size {
		similarResult = similarResult[:size]
	}

	for _, r := range similarResult {
		result = append(result, r.Get("city").(*RegionCity))
	}

	return
}
