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
	RegionProvinceStateEnabled  = 1 // 已启用
	RegionProvinceStateDisabled = 0 // 已禁用
)

type RegionProvinceDAO dbs.DAO

func NewRegionProvinceDAO() *RegionProvinceDAO {
	return dbs.NewDAO(&RegionProvinceDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionProvinces",
			Model:  new(RegionProvince),
			PkName: "id",
		},
	}).(*RegionProvinceDAO)
}

var SharedRegionProvinceDAO *RegionProvinceDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionProvinceDAO = NewRegionProvinceDAO()
	})
}

// EnableRegionProvince 启用条目
func (this *RegionProvinceDAO) EnableRegionProvince(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Attr("valueId", id).
		Set("state", RegionProvinceStateEnabled).
		Update()
	return err
}

// DisableRegionProvince 禁用条目
func (this *RegionProvinceDAO) DisableRegionProvince(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Attr("valueId", id).
		Set("state", RegionProvinceStateDisabled).
		Update()
	return err
}

// FindEnabledRegionProvince 查找启用中的条目
func (this *RegionProvinceDAO) FindEnabledRegionProvince(tx *dbs.Tx, id int64) (*RegionProvince, error) {
	result, err := this.Query(tx).
		Attr("valueId", id).
		Attr("state", RegionProvinceStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionProvince), err
}

// FindRegionProvinceName 根据主键查找名称
func (this *RegionProvinceDAO) FindRegionProvinceName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Attr("valueId", id).
		Result("name").
		FindStringCol("")
}

// FindProvinceIdWithDataId 根据数据ID查找省份
func (this *RegionProvinceDAO) FindProvinceIdWithDataId(tx *dbs.Tx, dataId string) (int64, error) {
	return this.Query(tx).
		Attr("dataId", dataId).
		Result(RegionProvinceField_ValueId).
		FindInt64Col(0)
}

// FindProvinceIdWithName 根据省份名查找省份ID
func (this *RegionProvinceDAO) FindProvinceIdWithName(tx *dbs.Tx, countryId int64, provinceName string) (int64, error) {
	return this.Query(tx).
		Attr("countryId", countryId).
		Where("(name=:provinceName OR customName=:provinceName OR JSON_CONTAINS(codes, :provinceNameJSON) OR JSON_CONTAINS(customCodes, :provinceNameJSON))").
		Param("provinceName", provinceName).
		Param("provinceNameJSON", strconv.Quote(provinceName)). // 查询的需要是个JSON字符串，所以这里加双引号
		Result(RegionProvinceField_ValueId).
		FindInt64Col(0)
}

// CreateProvince 创建省份
func (this *RegionProvinceDAO) CreateProvince(tx *dbs.Tx, countryId int64, name string, dataId string) (int64, error) {
	var op = NewRegionProvinceOperator()
	op.CountryId = countryId
	op.Name = name
	op.DataId = dataId
	op.State = RegionProvinceStateEnabled

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
	var provinceId = types.Int64(op.Id)

	err = this.Query(tx).
		Pk(provinceId).
		Set(RegionProvinceField_ValueId, provinceId).
		UpdateQuickly()
	if err != nil {
		return 0, err
	}

	return provinceId, nil
}

// FindAllEnabledProvincesWithCountryId 查找某个国家/地区的所有省份
func (this *RegionProvinceDAO) FindAllEnabledProvincesWithCountryId(tx *dbs.Tx, countryId int64) (result []*RegionProvince, err error) {
	_, err = this.Query(tx).
		State(RegionProvinceStateEnabled).
		Attr("countryId", countryId).
		Asc(RegionProvinceField_ValueId).
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledProvinces 查找所有省份
func (this *RegionProvinceDAO) FindAllEnabledProvinces(tx *dbs.Tx) (result []*RegionProvince, err error) {
	_, err = this.Query(tx).
		State(RegionProvinceStateEnabled).
		Asc(RegionProvinceField_ValueId).
		Slice(&result).
		FindAll()
	return
}

// UpdateProvinceCustom 修改自定义省份信息
func (this *RegionProvinceDAO) UpdateProvinceCustom(tx *dbs.Tx, provinceId int64, customName string, customCodes []string) error {
	if customCodes == nil {
		customCodes = []string{}
	}
	customCodesJSON, err := json.Marshal(customCodes)
	if err != nil {
		return err
	}

	return this.Query(tx).
		Attr("valueId", provinceId).
		Set("customName", customName).
		Set("customCodes", customCodesJSON).
		UpdateQuickly()
}

// FindSimilarProvinces 查找类似省份名
func (this *RegionProvinceDAO) FindSimilarProvinces(provinces []*RegionProvince, provinceName string, size int) (result []*RegionProvince) {
	if len(provinces) == 0 {
		return
	}

	var similarResult = []maps.Map{}

	for _, province := range provinces {
		var similarityList = []float32{}
		for _, code := range province.AllCodes() {
			var similarity = utils.Similar(provinceName, code)
			if similarity > 0 {
				similarityList = append(similarityList, similarity)
			}
		}
		if len(similarityList) > 0 {
			similarResult = append(similarResult, maps.Map{
				"similarity": numberutils.Max(similarityList...),
				"province":   province,
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
		result = append(result, r.Get("province").(*RegionProvince))
	}

	return
}
