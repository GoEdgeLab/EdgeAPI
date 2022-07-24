package models

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"strconv"
)

const (
	ClientBrowserStateEnabled  = 1 // 已启用
	ClientBrowserStateDisabled = 0 // 已禁用
)

var clientBrowserNameAndIdCacheMap = map[string]int64{}

type ClientBrowserDAO dbs.DAO

func NewClientBrowserDAO() *ClientBrowserDAO {
	return dbs.NewDAO(&ClientBrowserDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeClientBrowsers",
			Model:  new(ClientBrowser),
			PkName: "id",
		},
	}).(*ClientBrowserDAO)
}

var SharedClientBrowserDAO *ClientBrowserDAO

func init() {
	dbs.OnReady(func() {
		SharedClientBrowserDAO = NewClientBrowserDAO()
	})
}

// EnableClientBrowser 启用条目
func (this *ClientBrowserDAO) EnableClientBrowser(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ClientBrowserStateEnabled).
		Update()
	return err
}

// DisableClientBrowser 禁用条目
func (this *ClientBrowserDAO) DisableClientBrowser(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ClientBrowserStateDisabled).
		Update()
	return err
}

// FindEnabledClientBrowser 查找启用中的条目
func (this *ClientBrowserDAO) FindEnabledClientBrowser(tx *dbs.Tx, id int64) (*ClientBrowser, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ClientBrowserStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ClientBrowser), err
}

// FindClientBrowserName 根据主键查找名称
func (this *ClientBrowserDAO) FindClientBrowserName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindBrowserIdWithNameCacheable 根据浏览器名称查找浏览器ID
func (this *ClientBrowserDAO) FindBrowserIdWithNameCacheable(tx *dbs.Tx, browserName string) (int64, error) {
	SharedCacheLocker.RLock()
	browserId, ok := clientBrowserNameAndIdCacheMap[browserName]
	if ok {
		SharedCacheLocker.RUnlock()
		return browserId, nil
	}
	SharedCacheLocker.RUnlock()

	browserId, err := this.Query(tx).
		Where("JSON_CONTAINS(codes, :browserName)").
		Param("browserName", strconv.Quote(browserName)). // 查询的需要是个JSON字符串，所以这里加双引号
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}

	if browserId > 0 {
		// 只有找到的时候才放入缓存，以便于我们可以在不存在的时候创建一条新的记录
		SharedCacheLocker.Lock()
		clientBrowserNameAndIdCacheMap[browserName] = browserId
		SharedCacheLocker.Unlock()
	}

	return browserId, nil
}

// CreateBrowser 创建浏览器
func (this *ClientBrowserDAO) CreateBrowser(tx *dbs.Tx, browserName string) (int64, error) {
	var maxlength = 50
	if len(browserName) > maxlength {
		browserName = browserName[:50]
	}

	SharedCacheLocker.Lock()
	defer SharedCacheLocker.Unlock()

	// 检查是否已经创建
	browserId, err := this.Query(tx).
		Attr("name", browserName).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return 0, err
	}
	if browserId > 0 {
		return browserId, nil
	}

	var op = NewClientBrowserOperator()
	op.Name = browserName
	codes := []string{browserName}
	codesJSON, err := json.Marshal(codes)
	if err != nil {
		return 0, err
	}
	op.Codes = codesJSON
	op.State = ClientBrowserStateEnabled
	return this.SaveInt64(tx, op)
}
