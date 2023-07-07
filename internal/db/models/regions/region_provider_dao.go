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
	RegionProviderStateEnabled  = 1 // 已启用
	RegionProviderStateDisabled = 0 // 已禁用
)

type RegionProviderDAO dbs.DAO

func NewRegionProviderDAO() *RegionProviderDAO {
	return dbs.NewDAO(&RegionProviderDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeRegionProviders",
			Model:  new(RegionProvider),
			PkName: "id",
		},
	}).(*RegionProviderDAO)
}

var SharedRegionProviderDAO *RegionProviderDAO

func init() {
	dbs.OnReady(func() {
		SharedRegionProviderDAO = NewRegionProviderDAO()
	})
}

// EnableRegionProvider 启用条目
func (this *RegionProviderDAO) EnableRegionProvider(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Attr("valueId", id).
		Set("state", RegionProviderStateEnabled).
		Update()
	return err
}

// DisableRegionProvider 禁用条目
func (this *RegionProviderDAO) DisableRegionProvider(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Attr("valueId", id).
		Set("state", RegionProviderStateDisabled).
		Update()
	return err
}

// FindEnabledRegionProvider 查找启用中的条目
func (this *RegionProviderDAO) FindEnabledRegionProvider(tx *dbs.Tx, id int64) (*RegionProvider, error) {
	result, err := this.Query(tx).
		Attr("valueId", id).
		Attr("state", RegionProviderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionProvider), err
}

// FindRegionProviderName 根据主键查找名称
func (this *RegionProviderDAO) FindRegionProviderName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Attr("valueId", id).
		Result("name").
		FindStringCol("")
}

// FindProviderIdWithName 根据服务商名称查找服务商ID
func (this *RegionProviderDAO) FindProviderIdWithName(tx *dbs.Tx, providerName string) (int64, error) {
	return this.Query(tx).
		Where("(name=:providerName OR customName=:providerName OR JSON_CONTAINS(codes, :providerNameJSON) OR JSON_CONTAINS(customCodes, :providerNameJSON))").
		Param("providerName", providerName).
		Param("providerNameJSON", strconv.Quote(providerName)). // 查询的需要是个JSON字符串，所以这里加双引号
		Result(RegionProviderField_ValueId).
		FindInt64Col(0)
}

// CreateProvider 创建Provider
func (this *RegionProviderDAO) CreateProvider(tx *dbs.Tx, name string) (int64, error) {
	var op = NewRegionProviderOperator()
	op.Name = name

	codesJSON, err := json.Marshal([]string{name})
	if err != nil {
		return 0, err
	}
	op.Codes = codesJSON
	providerId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = this.Query(tx).
		Pk(providerId).
		Set(RegionProviderField_ValueId, providerId).
		UpdateQuickly()
	if err != nil {
		return 0, err
	}

	return providerId, nil
}

// FindAllEnabledProviders 查找所有服务商
func (this *RegionProviderDAO) FindAllEnabledProviders(tx *dbs.Tx) (result []*RegionProvider, err error) {
	_, err = this.Query(tx).
		State(RegionProviderStateEnabled).
		Slice(&result).
		FindAll()
	return
}

// UpdateProviderCustom 修改ISP自定义信息
func (this *RegionProviderDAO) UpdateProviderCustom(tx *dbs.Tx, providerId int64, customName string, customCodes []string) error {
	if customCodes == nil {
		customCodes = []string{}
	}
	customCodesJSON, err := json.Marshal(customCodes)
	if err != nil {
		return err
	}

	return this.Query(tx).
		Attr("valueId", providerId).
		Set("customName", customName).
		Set("customCodes", customCodesJSON).
		UpdateQuickly()
}

// FindSimilarProviders 查找类似ISP运营商
func (this *RegionProviderDAO) FindSimilarProviders(providers []*RegionProvider, providerName string, size int) (result []*RegionProvider) {
	if len(providers) == 0 {
		return
	}

	var similarResult = []maps.Map{}

	for _, provider := range providers {
		var similarityList = []float32{}
		for _, code := range provider.AllCodes() {
			var similarity = utils.Similar(providerName, code)
			if similarity > 0 {
				similarityList = append(similarityList, similarity)
			}
		}
		if len(similarityList) > 0 {
			similarResult = append(similarResult, maps.Map{
				"similarity": numberutils.Max(similarityList...),
				"provider":   provider,
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
		result = append(result, r.Get("provider").(*RegionProvider))
	}

	return
}
