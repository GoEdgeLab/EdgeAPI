package models

import (
	"encoding/base64"
	"github.com/TeaOSLab/EdgeAPI/internal/encrypt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strings"
)

const (
	DBNodeStateEnabled  = 1 // 已启用
	DBNodeStateDisabled = 0 // 已禁用
)

type DBNodeDAO dbs.DAO

const DBNodePasswordEncodedPrefix = "EDGE_ENCODED:"

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

var SharedDBNodeDAO *DBNodeDAO

func init() {
	dbs.OnReady(func() {
		SharedDBNodeDAO = NewDBNodeDAO()
	})
}

// 启用条目
func (this *DBNodeDAO) EnableDBNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", DBNodeStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *DBNodeDAO) DisableDBNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", DBNodeStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *DBNodeDAO) FindEnabledDBNode(tx *dbs.Tx, id int64) (*DBNode, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", DBNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	node := result.(*DBNode)
	node.Password = this.DecodePassword(node.Password)
	return node, nil
}

// 根据主键查找名称
func (this *DBNodeDAO) FindDBNodeName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 计算可用的节点数量
func (this *DBNodeDAO) CountAllEnabledNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(DBNodeStateEnabled).
		Count()
}

// 获取单页的节点
func (this *DBNodeDAO) ListEnabledNodes(tx *dbs.Tx, offset int64, size int64) (result []*DBNode, err error) {
	_, err = this.Query(tx).
		State(DBNodeStateEnabled).
		Offset(offset).
		Limit(size).
		Slice(&result).
		DescPk().
		FindAll()
	for _, node := range result {
		node.Password = this.DecodePassword(node.Password)
	}
	return
}

// 创建节点
func (this *DBNodeDAO) CreateDBNode(tx *dbs.Tx, isOn bool, name string, description string, host string, port int32, database string, username string, password string, charset string) (int64, error) {
	op := NewDBNodeOperator()
	op.State = NodeStateEnabled
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	op.Host = host
	op.Port = port
	op.Database = database
	op.Username = username
	op.Password = this.EncodePassword(password)
	op.Charset = charset
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改节点
func (this *DBNodeDAO) UpdateNode(tx *dbs.Tx, nodeId int64, isOn bool, name string, description string, host string, port int32, database string, username string, password string, charset string) error {
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
	op.Password = this.EncodePassword(password)
	op.Charset = charset
	err := this.Save(tx, op)
	return err
}

// 查找所有可用的数据库节点
func (this *DBNodeDAO) FindAllEnabledAndOnDBNodes(tx *dbs.Tx) (result []*DBNode, err error) {
	_, err = this.Query(tx).
		State(DBNodeStateEnabled).
		Attr("isOn", true).
		Slice(&result).
		DescPk().
		FindAll()
	for _, node := range result {
		node.Password = this.DecodePassword(node.Password)
	}
	return
}

// 加密密码
func (this *DBNodeDAO) EncodePassword(password string) string {
	if strings.HasPrefix(password, DBNodePasswordEncodedPrefix) {
		return password
	}
	encodedString := base64.StdEncoding.EncodeToString(encrypt.MagicKeyEncode([]byte(password)))
	return DBNodePasswordEncodedPrefix + encodedString
}

// 解密密码
func (this *DBNodeDAO) DecodePassword(password string) string {
	if !strings.HasPrefix(password, DBNodePasswordEncodedPrefix) {
		return password
	}
	dataString := password[len(DBNodePasswordEncodedPrefix):]
	data, err := base64.StdEncoding.DecodeString(dataString)
	if err != nil {
		return password
	}
	return string(encrypt.MagicKeyDecode(data))
}
