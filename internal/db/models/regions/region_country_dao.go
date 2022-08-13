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
	"github.com/mozillazg/go-pinyin"
	"sort"
	"strconv"
	"strings"
)

const (
	RegionCountryStateEnabled  = 1 // 已启用
	RegionCountryStateDisabled = 0 // 已禁用
)

const (
	CountryChinaId = 1
)

var regionCountryNameAndIdCacheMap = map[string]int64{} // country name => id
var regionCountryIdAndNameCacheMap = map[int64]string{} // country id => name

type RegionCountryDAO dbs.DAO

func NewRegionCountryDAO() *RegionCountryDAO {
	return dbs.NewDAO(&RegionCountryDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionCountries",
			Model:  new(RegionCountry),
			PkName: "id",
		},
	}).(*RegionCountryDAO)
}

var SharedRegionCountryDAO *RegionCountryDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionCountryDAO = NewRegionCountryDAO()
	})
}

// EnableRegionCountry 启用条目
func (this *RegionCountryDAO) EnableRegionCountry(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", RegionCountryStateEnabled).
		Update()
	return err
}

// DisableRegionCountry 禁用条目
func (this *RegionCountryDAO) DisableRegionCountry(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", RegionCountryStateDisabled).
		Update()
	return err
}

// FindEnabledRegionCountry 查找启用中的条目
func (this *RegionCountryDAO) FindEnabledRegionCountry(tx *dbs.Tx, id int64) (*RegionCountry, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", RegionCountryStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionCountry), err
}

// FindRegionCountryName 根据主键查找名称
func (this *RegionCountryDAO) FindRegionCountryName(tx *dbs.Tx, id int64) (string, error) {
	SharedCacheLocker.Lock()
	defer SharedCacheLocker.Unlock()

	name, ok := regionCountryIdAndNameCacheMap[id]
	if ok {
		return name, nil
	}

	name, err := this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
	if err != nil {
		return "", err
	}

	if len(name) > 0 {
		regionCountryIdAndNameCacheMap[id] = name
	}
	return name, nil
}

// FindCountryIdWithDataId 根据数据ID查找国家
func (this *RegionCountryDAO) FindCountryIdWithDataId(tx *dbs.Tx, dataId string) (int64, error) {
	return this.Query(tx).
		Attr("dataId", dataId).
		ResultPk().
		FindInt64Col(0)
}

// FindCountryIdWithName 根据国家名查找国家ID
func (this *RegionCountryDAO) FindCountryIdWithName(tx *dbs.Tx, countryName string) (int64, error) {
	return this.Query(tx).
		Where("(name=:countryName OR JSON_CONTAINS(codes, :countryNameJSON) OR customName=:countryName OR JSON_CONTAINS(customCodes, :countryNameJSON))").
		Param("countryName", countryName).
		Param("countryNameJSON", strconv.Quote(countryName)). // 查询的需要是个JSON字符串，所以这里加双引号
		ResultPk().
		FindInt64Col(0)
}

// FindCountryIdWithNameCacheable 根据国家名查找国家ID，并可使用缓存
func (this *RegionCountryDAO) FindCountryIdWithNameCacheable(tx *dbs.Tx, countryName string) (int64, error) {
	SharedCacheLocker.RLock()
	provinceId, ok := regionCountryNameAndIdCacheMap[countryName]
	if ok {
		SharedCacheLocker.RUnlock()
		return provinceId, nil
	}
	SharedCacheLocker.RUnlock()

	countryId, err := this.FindCountryIdWithName(tx, countryName)
	if err != nil {
		return 0, err
	}

	if countryId > 0 {
		SharedCacheLocker.Lock()
		regionCountryNameAndIdCacheMap[countryName] = countryId
		SharedCacheLocker.Unlock()
	}

	return countryId, nil
}

// CreateCountry 根据数据ID创建国家
func (this *RegionCountryDAO) CreateCountry(tx *dbs.Tx, name string, dataId string) (int64, error) {
	var op = NewRegionCountryOperator()
	op.Name = name

	pinyinPieces := pinyin.Pinyin(name, pinyin.NewArgs())
	pinyinResult := []string{}
	for _, piece := range pinyinPieces {
		pinyinResult = append(pinyinResult, strings.Join(piece, " "))
	}
	pinyinJSON, err := json.Marshal([]string{strings.Join(pinyinResult, " ")})
	op.Pinyin = pinyinJSON

	codes := []string{name}
	codesJSON, err := json.Marshal(codes)
	if err != nil {
		return 0, err
	}
	op.Codes = codesJSON

	op.DataId = dataId
	op.State = RegionCountryStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// FindAllEnabledCountriesOrderByPinyin 查找所有可用的国家并按拼音排序
func (this *RegionCountryDAO) FindAllEnabledCountriesOrderByPinyin(tx *dbs.Tx) (result []*RegionCountry, err error) {
	_, err = this.Query(tx).
		State(RegionCountryStateEnabled).
		Slice(&result).
		Asc("pinyin").
		FindAll()
	return
}

// FindAllCountries 查找所有可用的国家
func (this *RegionCountryDAO) FindAllCountries(tx *dbs.Tx) (result []*RegionCountry, err error) {
	_, err = this.Query(tx).
		State(RegionCountryStateEnabled).
		Slice(&result).
		AscPk().
		FindAll()
	return
}

// UpdateCountryCustom 修改国家/地区自定义
func (this *RegionCountryDAO) UpdateCountryCustom(tx *dbs.Tx, countryId int64, customName string, customCodes []string) error {
	if customCodes == nil {
		customCodes = []string{}
	}
	customCodesJSON, err := json.Marshal(customCodes)
	if err != nil {
		return err
	}

	defer func() {
		SharedCacheLocker.Lock()
		regionCountryNameAndIdCacheMap = map[string]int64{}
		regionCountryIdAndNameCacheMap = map[int64]string{}
		SharedCacheLocker.Unlock()
	}()

	return this.Query(tx).
		Pk(countryId).
		Set("customName", customName).
		Set("customCodes", customCodesJSON).
		UpdateQuickly()
}

// FindSimilarCountries 查找类似国家/地区名
func (this *RegionCountryDAO) FindSimilarCountries(countries []*RegionCountry, countryName string, size int) (result []*RegionCountry) {
	if len(countries) == 0 {
		return
	}

	var similarResult = []maps.Map{}

	for _, country := range countries {
		var similarityList = []float32{}
		for _, code := range country.AllCodes() {
			var similarity = utils.Similar(countryName, code)
			if similarity > 0 {
				similarityList = append(similarityList, similarity)
			}
		}
		if len(similarityList) > 0 {
			similarResult = append(similarResult, maps.Map{
				"similarity": numberutils.Max(similarityList...),
				"country":    country,
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
		result = append(result, r.Get("country").(*RegionCountry))
	}

	return
}
