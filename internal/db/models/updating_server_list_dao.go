package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"sort"
	"time"
)

type UpdatingServerListDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		var ticker = time.NewTicker(24 * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedUpdatingServerListDAO.CleanExpiredLists(nil, 7)
				if err != nil {
					remotelogs.Error("UpdatingServerListDAO", "CleanExpiredLists(): "+err.Error())
				}
			}
		})
	})
}

func NewUpdatingServerListDAO() *UpdatingServerListDAO {
	return dbs.NewDAO(&UpdatingServerListDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUpdatingServerLists",
			Model:  new(UpdatingServerList),
			PkName: "id",
		},
	}).(*UpdatingServerListDAO)
}

var SharedUpdatingServerListDAO *UpdatingServerListDAO

func init() {
	dbs.OnReady(func() {
		SharedUpdatingServerListDAO = NewUpdatingServerListDAO()
	})
}

// CreateList 创建待更新的服务列表
func (this *UpdatingServerListDAO) CreateList(tx *dbs.Tx, clusterId int64, serverIds []int64) error {
	if clusterId <= 0 || len(serverIds) == 0 {
		return nil
	}

	sort.Slice(serverIds, func(i, j int) bool {
		return serverIds[i] < serverIds[j]
	})

	serverIdsJSON, err := json.Marshal(serverIds)
	if err != nil {
		return err
	}

	var uniqueId = stringutil.Md5(types.String(clusterId) + "@" + string(serverIdsJSON))
	_, _, err = this.Query(tx).
		Set("uniqueId", uniqueId).
		Set("serverIds", serverIdsJSON).
		Set("clusterId", clusterId).
		Set("day", timeutil.Format("Ymd")).
		Replace() // 使用Replace，让ID可以自增
	return err
}

// FindLists 查找待更新服务列表
func (this *UpdatingServerListDAO) FindLists(tx *dbs.Tx, clusterIds []int64, lastId int64) (result []*UpdatingServerList, err error) {
	if len(clusterIds) == 0 {
		return
	}
	_, err = this.Query(tx).
		Attr("clusterId", clusterIds). // 即使clusterIds数量是变化的，这里也不需要使用Reuse(false)，因为clusterIds通常数量有限
		Gt("id", lastId).
		AscPk(). // 非常重要
		Slice(&result).
		FindAll()
	return
}

// FindLatestId 读取最新的ID
// 不需要区分集群
func (this *UpdatingServerListDAO) FindLatestId(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		ResultPk().
		DescPk().
		FindInt64Col(0)
}

// CleanExpiredLists 清除过期列表
func (this *UpdatingServerListDAO) CleanExpiredLists(tx *dbs.Tx, days int) error {
	if days <= 0 {
		days = 7
	}
	return this.Query(tx).
		Lt("day", timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))).
		DeleteQuickly()
}
