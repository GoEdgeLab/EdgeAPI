package models

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ClientSystemStateEnabled  = 1 // 已启用
	ClientSystemStateDisabled = 0 // 已禁用
)

var clientSystemNameAndIdCacheMap = map[string]int64{} // system name => id

type ClientSystemDAO dbs.DAO

func NewClientSystemDAO() *ClientSystemDAO {
	return dbs.NewDAO(&ClientSystemDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeClientSystems",
			Model:  new(ClientSystem),
			PkName: "id",
		},
	}).(*ClientSystemDAO)
}

var SharedClientSystemDAO *ClientSystemDAO

func init() {
	dbs.OnReady(func() {
		SharedClientSystemDAO = NewClientSystemDAO()
	})
}

// 启用条目
func (this *ClientSystemDAO) EnableClientSystem(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ClientSystemStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *ClientSystemDAO) DisableClientSystem(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ClientSystemStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *ClientSystemDAO) FindEnabledClientSystem(tx *dbs.Tx, id int64) (*ClientSystem, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ClientSystemStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ClientSystem), err
}

// 根据主键查找名称
func (this *ClientSystemDAO) FindClientSystemName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 根据操作系统名称查找系统ID
func (this *ClientSystemDAO) FindSystemIdWithNameCacheable(tx *dbs.Tx, systemName string) (int64, error) {
	SharedCacheLocker.RLock()
	systemId, ok := clientSystemNameAndIdCacheMap[systemName]
	if ok {
		SharedCacheLocker.RUnlock()
		return systemId, nil
	}
	SharedCacheLocker.RUnlock()

	systemId, err := this.Query(tx).
		Where("JSON_CONTAINS(codes, :systemName)").
		Param("systemName", "\""+systemName+"\""). // 查询的需要是个JSON字符串，所以这里加双引号
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}

	if systemId > 0 {
		// 只有找到的时候才放入缓存，以便于我们可以在不存在的时候创建一条新的记录
		SharedCacheLocker.Lock()
		clientSystemNameAndIdCacheMap[systemName] = systemId
		SharedCacheLocker.Unlock()
	}

	return systemId, nil
}

// 创建浏览器
func (this *ClientSystemDAO) CreateSystem(tx *dbs.Tx, systemName string) (int64, error) {
	op := NewClientSystemOperator()
	op.Name = systemName

	codes := []string{systemName}
	codesJSON, err := json.Marshal(codes)
	if err != nil {
		return 0, err
	}
	op.Codes = codesJSON

	op.State = ClientSystemStateEnabled
	return this.SaveInt64(tx, op)
}
