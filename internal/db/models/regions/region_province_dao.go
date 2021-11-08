package regions

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strconv"
)

const (
	RegionProvinceStateEnabled  = 1 // 已启用
	RegionProvinceStateDisabled = 0 // 已禁用
)

var regionProvinceNameAndIdCacheMap = map[string]int64{} // province name @ country id => province id

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
func (this *RegionProvinceDAO) EnableRegionProvince(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", RegionProvinceStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *RegionProvinceDAO) DisableRegionProvince(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", RegionProvinceStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *RegionProvinceDAO) FindEnabledRegionProvince(tx *dbs.Tx, id int64) (*RegionProvince, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", RegionProvinceStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionProvince), err
}

// 根据主键查找名称
func (this *RegionProvinceDAO) FindRegionProvinceName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 根据数据ID查找省份
func (this *RegionProvinceDAO) FindProvinceIdWithDataId(tx *dbs.Tx, dataId string) (int64, error) {
	return this.Query(tx).
		Attr("dataId", dataId).
		ResultPk().
		FindInt64Col(0)
}

// 根据省份名查找省份ID
func (this *RegionProvinceDAO) FindProvinceIdWithName(tx *dbs.Tx, countryId int64, provinceName string) (int64, error) {
	return this.Query(tx).
		Attr("countryId", countryId).
		Where("JSON_CONTAINS(codes, :provinceName)").
		Param("provinceName", strconv.Quote(provinceName)). // 查询的需要是个JSON字符串，所以这里加双引号
		ResultPk().
		FindInt64Col(0)
}

// 根据省份名查找省份ID，并可使用缓存
func (this *RegionProvinceDAO) FindProvinceIdWithNameCacheable(tx *dbs.Tx, countryId int64, provinceName string) (int64, error) {
	var key = provinceName + "@" + numberutils.FormatInt64(countryId)

	SharedCacheLocker.RLock()
	provinceId, ok := regionProvinceNameAndIdCacheMap[key]
	if ok {
		SharedCacheLocker.RUnlock()
		return provinceId, nil
	}
	SharedCacheLocker.RUnlock()

	provinceId, err := this.FindProvinceIdWithName(tx, countryId, provinceName)
	if err != nil {
		return 0, err
	}
	SharedCacheLocker.Lock()
	regionProvinceNameAndIdCacheMap[key] = provinceId
	SharedCacheLocker.Unlock()

	return provinceId, nil
}

// 创建省份
func (this *RegionProvinceDAO) CreateProvince(tx *dbs.Tx, countryId int64, name string, dataId string) (int64, error) {
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
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 查找所有省份
func (this *RegionProvinceDAO) FindAllEnabledProvincesWithCountryId(tx *dbs.Tx, countryId int64) (result []*RegionProvince, err error) {
	_, err = this.Query(tx).
		State(RegionProvinceStateEnabled).
		Attr("countryId", countryId).
		Asc().
		Slice(&result).
		FindAll()
	return
}
