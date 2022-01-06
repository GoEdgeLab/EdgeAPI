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
	RegionCityStateEnabled  = 1 // 已启用
	RegionCityStateDisabled = 0 // 已禁用
)

type RegionCityDAO dbs.DAO

var regionCityNameAndIdCacheMap = map[string]int64{} //  city name @ province id => id

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
		Pk(id).
		Set("state", RegionCityStateEnabled).
		Update()
	return err
}

// DisableRegionCity 禁用条目
func (this *RegionCityDAO) DisableRegionCity(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", RegionCityStateDisabled).
		Update()
	return err
}

// FindEnabledRegionCity 查找启用中的条目
func (this *RegionCityDAO) FindEnabledRegionCity(tx *dbs.Tx, id int64) (*RegionCity, error) {
	result, err := this.Query(tx).
		Pk(id).
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
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindCityWithDataId 根据数据ID查找城市
func (this *RegionCityDAO) FindCityWithDataId(tx *dbs.Tx, dataId string) (int64, error) {
	return this.Query(tx).
		Attr("dataId", dataId).
		ResultPk().
		FindInt64Col(0)
}

// CreateCity 创建城市
func (this *RegionCityDAO) CreateCity(tx *dbs.Tx, provinceId int64, name string, dataId string) (int64, error) {
	op := NewRegionCityOperator()
	op.ProvinceId = provinceId
	op.Name = name
	op.DataId = dataId
	op.State = RegionCityStateEnabled

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

// FindCityIdWithNameCacheable 根据城市名查找城市ID
func (this *RegionCityDAO) FindCityIdWithNameCacheable(tx *dbs.Tx, provinceId int64, cityName string) (int64, error) {
	key := cityName + "@" + numberutils.FormatInt64(provinceId)

	SharedCacheLocker.RLock()
	cityId, ok := regionCityNameAndIdCacheMap[key]
	if ok {
		SharedCacheLocker.RUnlock()
		return cityId, nil
	}
	SharedCacheLocker.RUnlock()

	cityId, err := this.Query(tx).
		Attr("provinceId", provinceId).
		Where("JSON_CONTAINS(codes, :cityName)").
		Param("cityName", strconv.Quote(cityName)). // 查询的需要是个JSON字符串，所以这里加双引号
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}
	SharedCacheLocker.Lock()
	regionCityNameAndIdCacheMap[key] = cityId
	SharedCacheLocker.Unlock()

	return cityId, nil
}

// FindAllEnabledCities 获取所有城市信息
func (this *RegionCityDAO) FindAllEnabledCities(tx *dbs.Tx) (result []*RegionCity, err error) {
	_, err = this.Query(tx).
		State(RegionCityStateEnabled).
		Slice(&result).
		FindAll()
	return
}
