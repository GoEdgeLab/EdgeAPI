package clients

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/ttlcache"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

const (
	ClientBrowserStateEnabled  = 1 // 已启用
	ClientBrowserStateDisabled = 0 // 已禁用
)

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedClientBrowserDAO.Clean(nil, 7) // 只保留N天
				if err != nil {
					remotelogs.Error("SharedClientBrowserDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

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

// CreateBrowserIfNotExists 创建浏览器信息
func (this *ClientBrowserDAO) CreateBrowserIfNotExists(tx *dbs.Tx, browserName string) error {
	const maxlength = 50
	if len(browserName) > maxlength {
		browserName = browserName[:50]
	}

	// 检查缓存
	var cacheKey = "clientBrowser:" + browserName
	var cacheItem = ttlcache.SharedCache.Read(cacheKey)
	if cacheItem != nil {
		return nil
	}

	// 检查是否已经存在
	// 不需要加状态条件
	browserId, err := this.Query(tx).
		Attr("name", browserName).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if browserId > 0 {
		// 加入缓存，但缓存时间不要过长，因为有别的操作在更新数据
		ttlcache.SharedCache.Write(cacheKey, browserId, time.Now().Unix()+3600)

		return this.Query(tx).
			Pk(browserId).
			Set("createdDay", timeutil.Format("Ymd")).
			UpdateQuickly()
	}

	// 如果不存在，则创建之
	var op = NewClientBrowserOperator()
	op.Name = browserName
	op.CreatedDay = timeutil.Format("Ymd")
	op.State = ClientBrowserStateEnabled
	browserId, err = this.SaveInt64(tx, op)
	if err != nil && models.CheckSQLErrCode(err, 1062 /** duplicate entry **/) {
		return nil
	}

	// 加入缓存，但缓存时间不要过长，因为有别的操作在更新数据
	if browserId > 0 {
		ttlcache.SharedCache.Write(cacheKey, browserId, time.Now().Unix()+3600)
	}

	return err
}

// Clean 清理
func (this *ClientBrowserDAO) Clean(tx *dbs.Tx, days int) error {
	if days <= 0 {
		days = 30
	}

	return this.Query(tx).
		Lt("createdDay", timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))).
		DeleteQuickly()
}
