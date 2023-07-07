package regions

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"sort"
	"strconv"
)

const (
	RegionTownStateEnabled  = 1 // 已启用
	RegionTownStateDisabled = 0 // 已禁用
)

type RegionTownDAO dbs.DAO

func NewRegionTownDAO() *RegionTownDAO {
	return dbs.NewDAO(&RegionTownDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionTowns",
			Model:  new(RegionTown),
			PkName: "id",
		},
	}).(*RegionTownDAO)
}

var SharedRegionTownDAO *RegionTownDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionTownDAO = NewRegionTownDAO()
	})
}

// EnableRegionTown 启用条目
func (this *RegionTownDAO) EnableRegionTown(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Attr("valueId", id).
		Set("state", RegionTownStateEnabled).
		Update()
	return err
}

// DisableRegionTown 禁用条目
func (this *RegionTownDAO) DisableRegionTown(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Attr("valueId", id).
		Set("state", RegionTownStateDisabled).
		Update()
	return err
}

// FindEnabledRegionTown 查找启用中的区县
func (this *RegionTownDAO) FindEnabledRegionTown(tx *dbs.Tx, id int64) (*RegionTown, error) {
	result, err := this.Query(tx).
		Attr("valueId", id).
		Attr("state", RegionTownStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionTown), err
}

// FindAllRegionTowns 获取所有的区县
func (this *RegionTownDAO) FindAllRegionTowns(tx *dbs.Tx) (result []*RegionTown, err error) {
	_, err = this.Query(tx).
		State(RegionTownStateEnabled).
		Asc(RegionTownField_ValueId).
		Slice(&result).
		FindAll()
	return
}

// FindAllRegionTownsWithCityId 根据城市查找区县
func (this *RegionTownDAO) FindAllRegionTownsWithCityId(tx *dbs.Tx, cityId int64) (result []*RegionTown, err error) {
	_, err = this.Query(tx).
		State(RegionTownStateEnabled).
		Attr("cityId", cityId).
		Asc(RegionTownField_ValueId).
		Slice(&result).
		FindAll()
	return
}

// FindTownIdWithName 根据区县名查找区县ID
func (this *RegionTownDAO) FindTownIdWithName(tx *dbs.Tx, cityId int64, townName string) (int64, error) {
	return this.Query(tx).
		Attr("cityId", cityId).
		Where("(name=:townName OR customName=:townName OR JSON_CONTAINS(codes, :townNameJSON) OR JSON_CONTAINS(customCodes, :townNameJSON))").
		Param("townName", townName).
		Param("townNameJSON", strconv.Quote(townName)). // 查询的需要是个JSON字符串，所以这里加双引号
		Result(RegionTownField_ValueId).
		FindInt64Col(0)
}

// FindRegionTownName 根据主键查找名称
func (this *RegionTownDAO) FindRegionTownName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Attr("valueId", id).
		Result("name").
		FindStringCol("")
}

// UpdateTownCustom 修改自定义县级信息
func (this *RegionTownDAO) UpdateTownCustom(tx *dbs.Tx, townId int64, customName string, customCodes []string) error {
	if customCodes == nil {
		customCodes = []string{}
	}
	customCodesJSON, err := json.Marshal(customCodes)
	if err != nil {
		return err
	}
	return this.Query(tx).
		Attr("valueId", townId).
		Set("customName", customName).
		Set("customCodes", customCodesJSON).
		UpdateQuickly()
}

// FindSimilarTowns 查找类似区县
func (this *RegionTownDAO) FindSimilarTowns(towns []*RegionTown, townName string, size int) (result []*RegionTown) {
	if len(towns) == 0 {
		return
	}

	var similarResult = []maps.Map{}

	for _, town := range towns {
		var similarityList = []float32{}
		for _, code := range town.AllCodes() {
			var similarity = utils.Similar(townName, code)
			if similarity > 0 {
				similarityList = append(similarityList, similarity)
			}
		}
		if len(similarityList) > 0 {
			similarResult = append(similarResult, maps.Map{
				"similarity": numberutils.Max(similarityList...),
				"town":       town,
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
		result = append(result, r.Get("town").(*RegionTown))
	}

	return
}

// CreateTown 创建区县
func (this *RegionTownDAO) CreateTown(tx *dbs.Tx, cityId int64, townName string) (int64, error) {
	var op = NewRegionTownOperator()
	op.CityId = cityId
	op.Name = townName

	codes, err := json.Marshal([]string{townName})
	if err != nil {
		return 0, err
	}
	op.Codes = codes

	op.State = RegionTownStateEnabled
	townId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = this.Query(tx).
		Pk(townId).
		Set(RegionTownField_ValueId, townId).
		UpdateQuickly()
	if err != nil {
		return 0, err
	}

	return townId, nil
}
