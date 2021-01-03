package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"time"
)

const (
	IPItemStateEnabled  = 1 // 已启用
	IPItemStateDisabled = 0 // 已禁用
)

type IPItemDAO dbs.DAO

func NewIPItemDAO() *IPItemDAO {
	return dbs.NewDAO(&IPItemDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeIPItems",
			Model:  new(IPItem),
			PkName: "id",
		},
	}).(*IPItemDAO)
}

var SharedIPItemDAO *IPItemDAO

func init() {
	dbs.OnReady(func() {
		SharedIPItemDAO = NewIPItemDAO()
	})
}

// 启用条目
func (this *IPItemDAO) EnableIPItem(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", IPItemStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *IPItemDAO) DisableIPItem(tx *dbs.Tx, id int64) error {
	version, err := SharedIPListDAO.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(id).
		Set("state", IPItemStateDisabled).
		Set("version", version).
		Update()
	return err
}

// 查找启用中的条目
func (this *IPItemDAO) FindEnabledIPItem(tx *dbs.Tx, id int64) (*IPItem, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", IPItemStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*IPItem), err
}

// 创建IP
func (this *IPItemDAO) CreateIPItem(tx *dbs.Tx, listId int64, ipFrom string, ipTo string, expiredAt int64, reason string) (int64, error) {
	version, err := SharedIPListDAO.IncreaseVersion(tx)
	if err != nil {
		return 0, err
	}

	op := NewIPItemOperator()
	op.ListId = listId
	op.IpFrom = ipFrom
	op.IpTo = ipTo
	op.Reason = reason
	op.Version = version
	if expiredAt < 0 {
		expiredAt = 0
	}
	op.ExpiredAt = expiredAt
	op.State = IPItemStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改IP
func (this *IPItemDAO) UpdateIPItem(tx *dbs.Tx, itemId int64, ipFrom string, ipTo string, expiredAt int64, reason string) error {
	if itemId <= 0 {
		return errors.New("invalid itemId")
	}

	listId, err := this.Query(tx).
		Pk(itemId).
		Result("listId").
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if listId == 0 {
		return errors.New("not found")
	}

	version, err := SharedIPListDAO.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	op := NewIPItemOperator()
	op.Id = itemId
	op.IpFrom = ipFrom
	op.IpTo = ipTo
	op.Reason = reason
	if expiredAt < 0 {
		expiredAt = 0
	}
	op.ExpiredAt = expiredAt
	op.Version = version
	err = this.Save(tx, op)
	return err
}

// 计算IP数量
func (this *IPItemDAO) CountIPItemsWithListId(tx *dbs.Tx, listId int64) (int64, error) {
	return this.Query(tx).
		State(IPItemStateEnabled).
		Attr("listId", listId).
		Count()
}

// 查找IP列表
func (this *IPItemDAO) ListIPItemsWithListId(tx *dbs.Tx, listId int64, offset int64, size int64) (result []*IPItem, err error) {
	_, err = this.Query(tx).
		State(IPItemStateEnabled).
		Attr("listId", listId).
		DescPk().
		Slice(&result).
		Offset(offset).
		Limit(size).
		FindAll()
	return
}

// 根据版本号查找IP列表
func (this *IPItemDAO) ListIPItemsAfterVersion(tx *dbs.Tx, version int64, size int64) (result []*IPItem, err error) {
	_, err = this.Query(tx).
		// 这里不要设置状态参数，因为我们要知道哪些是删除的
		Gt("version", version).
		Where("(expiredAt=0 OR expiredAt>:expiredAt)").
		Param("expiredAt", time.Now().Unix()).
		Asc("version").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// 查找IPItem对应的列表ID
func (this *IPItemDAO) FindItemListId(tx *dbs.Tx, itemId int64) (int64, error) {
	return this.Query(tx).
		Pk(itemId).
		Result("listId").
		FindInt64Col(0)
}
