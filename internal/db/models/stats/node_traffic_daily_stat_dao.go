package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type NodeTrafficDailyStatDAO dbs.DAO

func NewNodeTrafficDailyStatDAO() *NodeTrafficDailyStatDAO {
	return dbs.NewDAO(&NodeTrafficDailyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeTrafficDailyStats",
			Model:  new(NodeTrafficDailyStat),
			PkName: "id",
		},
	}).(*NodeTrafficDailyStatDAO)
}

var SharedNodeTrafficDailyStatDAO *NodeTrafficDailyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeTrafficDailyStatDAO = NewNodeTrafficDailyStatDAO()
	})
}

// 增加流量
func (this *NodeTrafficDailyStatDAO) IncreaseDailyBytes(tx *dbs.Tx, nodeId int64, day string, bytes int64) error {
	if len(day) != 8 {
		return errors.New("invalid day '" + day + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		InsertOrUpdateQuickly(maps.Map{
			"nodeId": nodeId,
			"day":    day,
			"bytes":  bytes,
		}, maps.Map{
			"bytes": dbs.SQL("bytes+:bytes"),
		})
	if err != nil {
		return err
	}
	return nil
}
