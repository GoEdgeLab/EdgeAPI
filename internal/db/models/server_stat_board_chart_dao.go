package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ServerStatBoardChartStateEnabled  = 1 // 已启用
	ServerStatBoardChartStateDisabled = 0 // 已禁用
)

type ServerStatBoardChartDAO dbs.DAO

func NewServerStatBoardChartDAO() *ServerStatBoardChartDAO {
	return dbs.NewDAO(&ServerStatBoardChartDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerStatBoardCharts",
			Model:  new(ServerStatBoardChart),
			PkName: "id",
		},
	}).(*ServerStatBoardChartDAO)
}

var SharedServerStatBoardChartDAO *ServerStatBoardChartDAO

func init() {
	dbs.OnReady(func() {
		SharedServerStatBoardChartDAO = NewServerStatBoardChartDAO()
	})
}

// EnableServerStatBoardChart 启用条目
func (this *ServerStatBoardChartDAO) EnableServerStatBoardChart(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ServerStatBoardChartStateEnabled).
		Update()
	return err
}

// DisableServerStatBoardChart 禁用条目
func (this *ServerStatBoardChartDAO) DisableServerStatBoardChart(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ServerStatBoardChartStateDisabled).
		Update()
	return err
}

// FindEnabledServerStatBoardChart 查找启用中的条目
func (this *ServerStatBoardChartDAO) FindEnabledServerStatBoardChart(tx *dbs.Tx, id uint64) (*ServerStatBoardChart, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ServerStatBoardChartStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ServerStatBoardChart), err
}

// EnableChart 启用图表
func (this *ServerStatBoardChartDAO) EnableChart(tx *dbs.Tx, boardId int64, chartId int64) error {
	op := NewServerStatBoardChartOperator()
	op.BoardId = boardId
	op.ChartId = chartId
	op.State = ServerStatBoardChartStateEnabled
	return this.Save(tx, op)
}

// DisableChart 禁用图表
func (this *ServerStatBoardChartDAO) DisableChart(tx *dbs.Tx, boardId int64, chartId int64) error {
	return this.Query(tx).
		Attr("borderId", boardId).
		Attr("chartId", chartId).
		Set("state", ServerStatBoardChartStateDisabled).
		UpdateQuickly()
}

// FindAllEnabledCharts 查找看板中所有图表
func (this *ServerStatBoardChartDAO) FindAllEnabledCharts(tx *dbs.Tx, boardId int64) (result []*ServerStatBoardChart, err error) {
	_, err = this.Query(tx).
		Attr("boardId", boardId).
		Desc("order").
		AscPk().
		State(ServerStatBoardChartStateEnabled).
		Slice(&result).
		FindAll()
	return
}
