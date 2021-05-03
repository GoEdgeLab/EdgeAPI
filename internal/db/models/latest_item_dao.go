package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"time"
)

type LatestItemType = string

const (
	LatestItemTypeCluster LatestItemType = "cluster"
	LatestItemTypeServer  LatestItemType = "server"
)

type LatestItemDAO dbs.DAO

func NewLatestItemDAO() *LatestItemDAO {
	return dbs.NewDAO(&LatestItemDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeLatestItems",
			Model:  new(LatestItem),
			PkName: "id",
		},
	}).(*LatestItemDAO)
}

var SharedLatestItemDAO *LatestItemDAO

func init() {
	dbs.OnReady(func() {
		SharedLatestItemDAO = NewLatestItemDAO()
	})
}

// IncreaseItemCount 增加数量
func (this *LatestItemDAO) IncreaseItemCount(tx *dbs.Tx, itemType LatestItemType, itemId int64) error {
	return this.Query(tx).
		InsertOrUpdateQuickly(maps.Map{
			"itemType":  itemType,
			"itemId":    itemId,
			"count":     1,
			"updatedAt": time.Now().Unix(),
		}, maps.Map{
			"count":     dbs.SQL("count+1"),
			"updatedAt": time.Now().Unix(),
		})
}
