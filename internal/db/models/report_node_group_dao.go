package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ReportNodeGroupStateEnabled  = 1 // 已启用
	ReportNodeGroupStateDisabled = 0 // 已禁用
)

type ReportNodeGroupDAO dbs.DAO

func NewReportNodeGroupDAO() *ReportNodeGroupDAO {
	return dbs.NewDAO(&ReportNodeGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeReportNodeGroups",
			Model:  new(ReportNodeGroup),
			PkName: "id",
		},
	}).(*ReportNodeGroupDAO)
}

var SharedReportNodeGroupDAO *ReportNodeGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedReportNodeGroupDAO = NewReportNodeGroupDAO()
	})
}

// EnableReportNodeGroup 启用条目
func (this *ReportNodeGroupDAO) EnableReportNodeGroup(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ReportNodeGroupStateEnabled).
		Update()
	return err
}

// DisableReportNodeGroup 禁用条目
func (this *ReportNodeGroupDAO) DisableReportNodeGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ReportNodeGroupStateDisabled).
		Update()
	return err
}

// FindEnabledReportNodeGroup 查找启用中的条目
func (this *ReportNodeGroupDAO) FindEnabledReportNodeGroup(tx *dbs.Tx, id int64) (*ReportNodeGroup, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ReportNodeGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ReportNodeGroup), err
}

// FindReportNodeGroupName 根据主键查找名称
func (this *ReportNodeGroupDAO) FindReportNodeGroupName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateGroup 创建
func (this *ReportNodeGroupDAO) CreateGroup(tx *dbs.Tx, name string) (int64, error) {
	var op = NewReportNodeGroupOperator()
	op.Name = name
	op.IsOn = true
	op.State = ReportNodeGroupStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateGroup 修改
func (this *ReportNodeGroupDAO) UpdateGroup(tx *dbs.Tx, groupId int64, name string) error {
	if groupId <= 0 {
		return errors.New("invalid groupId")
	}
	var op = NewReportNodeGroupOperator()
	op.Id = groupId
	op.Name = name
	return this.Save(tx, op)
}

// FindAllEnabledGroups 查找所有可用的分组
func (this *ReportNodeGroupDAO) FindAllEnabledGroups(tx *dbs.Tx) (result []*ReportNodeGroup, err error) {
	_, err = this.Query(tx).
		State(ReportNodeGroupStateEnabled).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledGroups 查找所有分组的数量
func (this *ReportNodeGroupDAO) CountAllEnabledGroups(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(ReportNodeGroupStateEnabled).
		Count()
}
