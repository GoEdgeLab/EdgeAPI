package models

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"strconv"
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

// EnableClientSystem 启用条目
func (this *ClientSystemDAO) EnableClientSystem(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ClientSystemStateEnabled).
		Update()
	return err
}

// DisableClientSystem 禁用条目
func (this *ClientSystemDAO) DisableClientSystem(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ClientSystemStateDisabled).
		Update()
	return err
}

// FindEnabledClientSystem 查找启用中的条目
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

// FindClientSystemName 根据主键查找名称
func (this *ClientSystemDAO) FindClientSystemName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindSystemIdWithNameCacheable 根据操作系统名称查找系统ID
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
		Param("systemName", strconv.Quote(systemName)). // 查询的需要是个JSON字符串，所以这里加双引号
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

// CreateSystem 创建浏览器
func (this *ClientSystemDAO) CreateSystem(tx *dbs.Tx, systemName string) (int64, error) {
	var maxlength = 50
	if len(systemName) > maxlength {
		systemName = systemName[:50]
	}

	SharedCacheLocker.Lock()
	defer SharedCacheLocker.Unlock()

	// 检查是否已经创建
	systemId, err := this.Query(tx).
		Attr("name", systemName).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}
	if systemId > 0 {
		return systemId, nil
	}

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
