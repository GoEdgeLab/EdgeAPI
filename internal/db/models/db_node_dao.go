package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	DBNodeStateEnabled  = 1 // 已启用
	DBNodeStateDisabled = 0 // 已禁用
)

type DBNodeDAO dbs.DAO

func NewDBNodeDAO() *DBNodeDAO {
	return dbs.NewDAO(&DBNodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeDBNodes",
			Model:  new(DBNode),
			PkName: "id",
		},
	}).(*DBNodeDAO)
}

var SharedDBNodeDAO = NewDBNodeDAO()

// 启用条目
func (this *DBNodeDAO) EnableDBNode(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", DBNodeStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *DBNodeDAO) DisableDBNode(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", DBNodeStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *DBNodeDAO) FindEnabledDBNode(id int64) (*DBNode, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", DBNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*DBNode), err
}

// 根据主键查找名称
func (this *DBNodeDAO) FindDBNodeName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 计算可用的节点数量
func (this *DBNodeDAO) CountAllEnabledNodes() (int64, error) {
	return this.Query().
		State(DBNodeStateEnabled).
		Count()
}

// 获取单页的节点
func (this *DBNodeDAO) ListEnabledNodes(offset int64, size int64) (result []*DBNode, err error) {
	_, err = this.Query().
		State(DBNodeStateEnabled).
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	return
}

// 创建节点
func (this *DBNodeDAO) CreateDBNode(isOn bool, name string, description string, host string, port int32, database string, username string, password string, charset string) (int64, error) {
	op := NewDBNodeOperator()
	op.State = NodeStateEnabled
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	op.Host = host
	op.Port = port
	op.Database = database
	op.Username = username
	op.Password = password
	op.Charset = charset
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改节点
func (this *DBNodeDAO) UpdateNode(nodeId int64, isOn bool, name string, description string, host string, port int32, database string, username string, password string, charset string) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	op := NewDBNodeOperator()
	op.Id = nodeId
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	op.Host = host
	op.Port = port
	op.Database = database
	op.Username = username
	op.Password = password
	op.Charset = charset
	_, err := this.Save(op)
	return err
}
