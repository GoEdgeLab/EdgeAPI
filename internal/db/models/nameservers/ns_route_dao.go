package nameservers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSRouteStateEnabled  = 1 // 已启用
	NSRouteStateDisabled = 0 // 已禁用
)

type NSRouteDAO dbs.DAO

func NewNSRouteDAO() *NSRouteDAO {
	return dbs.NewDAO(&NSRouteDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSRoutes",
			Model:  new(NSRoute),
			PkName: "id",
		},
	}).(*NSRouteDAO)
}

var SharedNSRouteDAO *NSRouteDAO

func init() {
	dbs.OnReady(func() {
		SharedNSRouteDAO = NewNSRouteDAO()
	})
}

// EnableNSRoute 启用条目
func (this *NSRouteDAO) EnableNSRoute(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSRouteStateEnabled).
		Update()
	return err
}

// DisableNSRoute 禁用条目
func (this *NSRouteDAO) DisableNSRoute(tx *dbs.Tx, routeId int64) error {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(routeId).
		Set("state", NSRouteStateDisabled).
		Set("version", version).
		Update()
	return err
}

// FindEnabledNSRoute 查找启用中的条目
func (this *NSRouteDAO) FindEnabledNSRoute(tx *dbs.Tx, id int64) (*NSRoute, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSRouteStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSRoute), err
}

// FindNSRouteName 根据主键查找名称
func (this *NSRouteDAO) FindNSRouteName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateRoute 创建线路
func (this *NSRouteDAO) CreateRoute(tx *dbs.Tx, clusterId int64, domainId int64, userId int64, name string, rangesJSON []byte) (int64, error) {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return 0, err
	}

	op := NewNSRouteOperator()
	op.ClusterId = clusterId
	op.DomainId = domainId
	op.UserId = userId
	op.Name = name
	if len(rangesJSON) > 0 {
		op.Ranges = rangesJSON
	} else {
		op.Ranges = "[]"
	}
	op.IsOn = true
	op.State = NSRouteStateEnabled
	op.Version = version
	return this.SaveInt64(tx, op)
}

// UpdateRoute 修改线路
func (this *NSRouteDAO) UpdateRoute(tx *dbs.Tx, routeId int64, name string, rangesJSON []byte) error {
	if routeId <= 0 {
		return errors.New("invalid routeId")
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	op := NewNSRouteOperator()
	op.Id = routeId
	op.Name = name
	if len(rangesJSON) > 0 {
		op.Ranges = rangesJSON
	} else {
		op.Ranges = "[]"
	}

	op.Version = version

	return this.Save(tx, op)
}

// UpdateRouteOrders 修改线路排序
func (this *NSRouteDAO) UpdateRouteOrders(tx *dbs.Tx, routeIds []int64) error {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	order := len(routeIds)
	for _, routeId := range routeIds {
		_, err = this.Query(tx).
			Pk(routeId).
			Set("order", order).
			Set("version", version).
			Update()
		if err != nil {
			return err
		}
		order--
	}
	return nil
}

// FindAllEnabledRoutes 列出所有线路
func (this *NSRouteDAO) FindAllEnabledRoutes(tx *dbs.Tx, clusterId int64, domainId int64, userId int64) (result []*NSRoute, err error) {
	query := this.Query(tx).
		State(NSRouteStateEnabled).
		Slice(&result).
		Desc("order").
		DescPk()
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	} else {
		// 不查询所有集群的线路
		query.Attr("clusterId", 0)
	}
	if domainId > 0 {
		query.Attr("domainId", domainId)
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}
	_, err = query.FindAll()
	return
}

// IncreaseVersion 增加版本
func (this *NSRouteDAO) IncreaseVersion(tx *dbs.Tx) (int64, error) {
	return models.SharedSysLockerDAO.Increase(tx, "NS_ROUTE_VERSION", 1)
}

// ListRoutesAfterVersion 列出某个版本后的域名
func (this *NSRouteDAO) ListRoutesAfterVersion(tx *dbs.Tx, version int64, size int64) (result []*NSRoute, err error) {
	if size <= 0 {
		size = 10000
	}

	_, err = this.Query(tx).
		Gte("version", version).
		Limit(size).
		Asc("version").
		Slice(&result).
		FindAll()
	return
}
