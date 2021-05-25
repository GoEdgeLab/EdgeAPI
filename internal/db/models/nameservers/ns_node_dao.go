package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSNodeStateEnabled  = 1 // 已启用
	NSNodeStateDisabled = 0 // 已禁用
)

type NSNodeDAO dbs.DAO

func NewNSNodeDAO() *NSNodeDAO {
	return dbs.NewDAO(&NSNodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSNodes",
			Model:  new(NSNode),
			PkName: "id",
		},
	}).(*NSNodeDAO)
}

var SharedNSNodeDAO *NSNodeDAO

func init() {
	dbs.OnReady(func() {
		SharedNSNodeDAO = NewNSNodeDAO()
	})
}

// EnableNSNode 启用条目
func (this *NSNodeDAO) EnableNSNode(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSNodeStateEnabled).
		Update()
	return err
}

// DisableNSNode 禁用条目
func (this *NSNodeDAO) DisableNSNode(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSNodeStateDisabled).
		Update()
	return err
}

// FindEnabledNSNode 查找启用中的条目
func (this *NSNodeDAO) FindEnabledNSNode(tx *dbs.Tx, id uint32) (*NSNode, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSNode), err
}
