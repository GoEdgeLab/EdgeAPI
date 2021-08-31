package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

type NodeIPAddressLogDAO dbs.DAO

func NewNodeIPAddressLogDAO() *NodeIPAddressLogDAO {
	return dbs.NewDAO(&NodeIPAddressLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeIPAddressLogs",
			Model:  new(NodeIPAddressLog),
			PkName: "id",
		},
	}).(*NodeIPAddressLogDAO)
}

var SharedNodeIPAddressLogDAO *NodeIPAddressLogDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeIPAddressLogDAO = NewNodeIPAddressLogDAO()
	})
}

// CreateLog 创建日志
func (this *NodeIPAddressLogDAO) CreateLog(tx *dbs.Tx, adminId int64, addrId int64, description string) error {
	addr, err := SharedNodeIPAddressDAO.FindEnabledAddress(tx, addrId)
	if err != nil {
		return err
	}
	if addr == nil {
		return nil
	}

	var op = NewNodeIPAddressLogOperator()
	op.AdminId = adminId
	op.AddressId = addrId
	op.Description = description
	op.CanAccess = addr.CanAccess
	op.IsOn = addr.IsOn
	op.IsUp = addr.IsUp
	op.Day = timeutil.Format("Ymd")
	return this.Save(tx, op)
}

// CountLogs 计算日志数量
func (this *NodeIPAddressLogDAO) CountLogs(tx *dbs.Tx, addrId int64) (int64, error) {
	var query = this.Query(tx)
	if addrId > 0 {
		query.Attr("addressId", addrId)
	}
	return query.Count()
}

// ListLogs 列出单页日志
func (this *NodeIPAddressLogDAO) ListLogs(tx *dbs.Tx, addrId int64, offset int64, size int64) (result []*NodeIPAddressLog, err error) {
	var query = this.Query(tx)
	if addrId > 0 {
		query.Attr("addressId", addrId)
	}
	_, err = query.Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
