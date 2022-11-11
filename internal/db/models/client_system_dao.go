package models

import (
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
	ClientSystemStateEnabled  = 1 // 已启用
	ClientSystemStateDisabled = 0 // 已禁用
)

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedClientSystemDAO.Clean(nil, 7) // 只保留N天
				if err != nil {
					remotelogs.Error("SharedClientSystemDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

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

// CreateSystemIfNotExists 创建系统信息
func (this *ClientSystemDAO) CreateSystemIfNotExists(tx *dbs.Tx, systemName string) error {
	const maxlength = 50
	if len(systemName) > maxlength {
		systemName = systemName[:50]
	}

	// 检查缓存
	var cacheKey = "clientSystem:" + systemName
	var cacheItem = ttlcache.SharedCache.Read(cacheKey)
	if cacheItem != nil {
		return nil
	}

	// 检查是否已经存在
	// 不需要加状态条件
	systemId, err := this.Query(tx).
		Attr("name", systemName).
		ResultPk().
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if systemId > 0 {
		// 加入缓存，但缓存时间不要过长，因为有别的操作在更新数据
		ttlcache.SharedCache.Write(cacheKey, systemId, time.Now().Unix()+3600)

		return this.Query(tx).
			Pk(systemId).
			Set("createdDay", timeutil.Format("Ymd")).
			UpdateQuickly()
	}

	var op = NewClientSystemOperator()
	op.Name = systemName
	op.CreatedDay = timeutil.Format("Ymd")
	op.State = ClientSystemStateEnabled
	systemId, err = this.SaveInt64(tx, op)
	if err != nil && CheckSQLErrCode(err, 1062 /** duplicate entry **/) {
		return nil
	}

	// 加入缓存，但缓存时间不要过长，因为有别的操作在更新数据
	if systemId > 0 {
		ttlcache.SharedCache.Write(cacheKey, systemId, time.Now().Unix()+3600)
	}

	return err
}

// Clean 清理
func (this *ClientSystemDAO) Clean(tx *dbs.Tx, days int) error {
	if days <= 0 {
		days = 30
	}

	return this.Query(tx).
		Lt("createdDay", timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))).
		DeleteQuickly()
}
