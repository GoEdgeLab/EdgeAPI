package models

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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

// 启用条目
func (this *RegionProvinceDAO) EnableRegionProvince(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionProvinceStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *RegionProvinceDAO) DisableRegionProvince(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", RegionProvinceStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *RegionProvinceDAO) FindEnabledRegionProvince(id int64) (*RegionProvince, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", RegionProvinceStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionProvince), err
}

// 根据主键查找名称
func (this *RegionProvinceDAO) FindRegionProvinceName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 根据数据ID查找省份
func (this *RegionProvinceDAO) FindProvinceIdWithDataId(dataId string) (int64, error) {
	return this.Query().
		Attr("dataId", dataId).
		ResultPk().
		FindInt64Col(0)
}

// 根据省份名查找省份ID
// TODO 加入缓存
func (this *RegionProvinceDAO) FindProvinceIdWithProvinceName(provinceName string) (int64, error) {
	return this.Query().
		Where("JSON_CONTAINS(codes, :provinceName)").
		Param("provinceName", "\""+provinceName+"\""). // 查询的需要是个JSON字符串，所以这里加双引号
		ResultPk().
		FindInt64Col(0)
}

// 创建省份
func (this *RegionProvinceDAO) CreateProvince(countryId int64, name string, dataId string) (int64, error) {
	op := NewRegionProvinceOperator()
	op.CountryId = countryId
	op.Name = name
	op.DataId = dataId
	op.State = RegionProvinceStateEnabled

	codes := []string{name}
	codesJSON, err := json.Marshal(codes)
	if err != nil {
		return 0, err
	}
	op.Codes = codesJSON
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 查找所有省份
func (this *RegionProvinceDAO) FindAllEnabledProvincesWithCountryId(countryId int64) (result []*RegionProvince, err error) {
	_, err = this.Query().
		State(RegionProvinceStateEnabled).
		Attr("countryId", countryId).
		Asc().
		Slice(&result).
		FindAll()
	return
}
