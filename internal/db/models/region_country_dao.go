package models

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"github.com/mozillazg/go-pinyin"
	"strings"
)

const (
	RegionCountryStateEnabled  = 1 // 已启用
	RegionCountryStateDisabled = 0 // 已禁用
)

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

// 启用条目
func (this *RegionCountryDAO) EnableRegionCountry(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionCountryStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *RegionCountryDAO) DisableRegionCountry(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionCountryStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *RegionCountryDAO) FindEnabledRegionCountry(id int64) (*RegionCountry, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", RegionCountryStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionCountry), err
}

// 根据主键查找名称
func (this *RegionCountryDAO) FindRegionCountryName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 根据数据ID查找国家
func (this *RegionCountryDAO) FindCountryIdWithDataId(dataId string) (int64, error) {
	return this.Query().
		Attr("dataId", dataId).
		ResultPk().
		FindInt64Col(0)
}

// 根据国家名查找国家ID
// TODO 加入缓存
func (this *RegionCountryDAO) FindCountryIdWithCountryName(countryName string) (int64, error) {
	return this.Query().
		Where("JSON_CONTAINS(codes, :countryName)").
		Param("countryName", "\""+countryName+"\""). // 查询的需要是个JSON字符串，所以这里加双引号
		ResultPk().
		FindInt64Col(0)
}

// 根据数据ID创建国家
func (this *RegionCountryDAO) CreateCountry(name string, dataId string) (int64, error) {
	op := NewRegionCountryOperator()
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
	err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 查找所有可用的国家
func (this *RegionCountryDAO) FindAllEnabledCountriesOrderByPinyin() (result []*RegionCountry, err error) {
	_, err = this.Query().
		State(RegionCountryStateEnabled).
		Slice(&result).
		Asc("pinyin").
		FindAll()
	return
}
