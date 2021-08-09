package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type NSQuestionOptionDAO dbs.DAO

func NewNSQuestionOptionDAO() *NSQuestionOptionDAO {
	return dbs.NewDAO(&NSQuestionOptionDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSQuestionOptions",
			Model:  new(NSQuestionOption),
			PkName: "id",
		},
	}).(*NSQuestionOptionDAO)
}

var SharedNSQuestionOptionDAO *NSQuestionOptionDAO

func init() {
	dbs.OnReady(func() {
		SharedNSQuestionOptionDAO = NewNSQuestionOptionDAO()
	})
}

// FindNSQuestionOptionName 根据主键查找名称
func (this *NSQuestionOptionDAO) FindNSQuestionOptionName(tx *dbs.Tx, id uint64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateOption 创建选项
func (this *NSQuestionOptionDAO) CreateOption(tx *dbs.Tx, name string, values maps.Map) (int64, error) {
	if values == nil {
		values = maps.Map{}
	}
	var op = NewNSQuestionOptionOperator()
	op.Name = name
	op.Values = values.AsJSON()
	return this.SaveInt64(tx, op)
}

// FindOption 读取选项
func (this *NSQuestionOptionDAO) FindOption(tx *dbs.Tx, optionId int64) (*NSQuestionOption, error) {
	one, err := this.Query(tx).
		Pk(optionId).
		Find()
	if one == nil {
		return nil, err
	}
	return one.(*NSQuestionOption), nil
}

// DeleteOption 删除选项
func (this *NSQuestionOptionDAO) DeleteOption(tx *dbs.Tx, optionId int64) error {
	_, err := this.Query(tx).
		Pk(optionId).
		Delete()
	return err
}
