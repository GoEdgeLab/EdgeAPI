package regions

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	RegionProviderStateEnabled  = 1 // 已启用
	RegionProviderStateDisabled = 0 // 已禁用
)

var regionProviderNameAndIdCacheMap = map[string]int64{} // provider name => id

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

// 启用条目
func (this *RegionProviderDAO) EnableRegionProvider(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", RegionProviderStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *RegionProviderDAO) DisableRegionProvider(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", RegionProviderStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *RegionProviderDAO) FindEnabledRegionProvider(tx *dbs.Tx, id uint32) (*RegionProvider, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", RegionProviderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*RegionProvider), err
}

// 根据主键查找名称
func (this *RegionProviderDAO) FindRegionProviderName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 根据服务商名称查找服务商ID
func (this *RegionProviderDAO) FindProviderIdWithNameCacheable(tx *dbs.Tx, providerName string) (int64, error) {
	SharedCacheLocker.RLock()
	providerId, ok := regionProviderNameAndIdCacheMap[providerName]
	if ok {
		SharedCacheLocker.RUnlock()
		return providerId, nil
	}
	SharedCacheLocker.RUnlock()

	providerId, err := this.Query(tx).
		Where("JSON_CONTAINS(codes, :providerName)").
		Param("providerName", "\""+providerName+"\""). // 查询的需要是个JSON字符串，所以这里加双引号
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}

	SharedCacheLocker.Lock()
	regionProviderNameAndIdCacheMap[providerName] = providerId
	SharedCacheLocker.Unlock()

	return providerId, nil
}

// 创建Provider
func (this *RegionProviderDAO) CreateProvider(tx *dbs.Tx, name string) (int64, error) {
	op := NewRegionProviderOperator()
	op.Name = name

	codesJSON, err := json.Marshal([]string{name})
	if err != nil {
		return 0, err
	}
	op.Codes = codesJSON
	return this.SaveInt64(tx, op)
}
