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
		Pk(id).
		Set("state", RegionTownStateEnabled).
		Update()
	return err
}

// DisableRegionTown 禁用条目
func (this *RegionTownDAO) DisableRegionTown(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", RegionTownStateDisabled).
		Update()
	return err
}

// FindEnabledRegionTown 查找启用中的条目
func (this *RegionTownDAO) FindEnabledRegionTown(tx *dbs.Tx, id uint32) (*RegionTown, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", RegionTownStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionTown), err
}

// FindRegionTownName 根据主键查找名称
func (this *RegionTownDAO) FindRegionTownName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
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
		Pk(townId).
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
