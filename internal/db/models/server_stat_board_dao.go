package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ServerStatBoardStateEnabled  = 1 // 已启用
	ServerStatBoardStateDisabled = 0 // 已禁用
)

type ServerStatBoardDAO dbs.DAO

func NewServerStatBoardDAO() *ServerStatBoardDAO {
	return dbs.NewDAO(&ServerStatBoardDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerStatBoards",
			Model:  new(ServerStatBoard),
			PkName: "id",
		},
	}).(*ServerStatBoardDAO)
}

var SharedServerStatBoardDAO *ServerStatBoardDAO

func init() {
	dbs.OnReady(func() {
		SharedServerStatBoardDAO = NewServerStatBoardDAO()
	})
}

// EnableServerStatBoard 启用条目
func (this *ServerStatBoardDAO) EnableServerStatBoard(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ServerStatBoardStateEnabled).
		Update()
	return err
}

// DisableServerStatBoard 禁用条目
func (this *ServerStatBoardDAO) DisableServerStatBoard(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ServerStatBoardStateDisabled).
		Update()
	return err
}

// FindEnabledServerStatBoard 查找启用中的条目
func (this *ServerStatBoardDAO) FindEnabledServerStatBoard(tx *dbs.Tx, id uint64) (*ServerStatBoard, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ServerStatBoardStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ServerStatBoard), err
}

// FindServerStatBoardName 根据主键查找名称
func (this *ServerStatBoardDAO) FindServerStatBoardName(tx *dbs.Tx, id uint64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindAllEnabledBoards 查找看板
func (this *ServerStatBoardDAO) FindAllEnabledBoards(tx *dbs.Tx, clusterId int64) (result []*ServerStatBoard, err error) {
	_, err = this.Query(tx).
		Attr("clusterId", clusterId).
		State(ServerStatBoardStateEnabled).
		Slice(&result).
		Desc("order").
		AscPk().
		FindAll()
	return
}
