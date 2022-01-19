package nameservers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSDomainStateEnabled  = 1 // 已启用
	NSDomainStateDisabled = 0 // 已禁用
)

type NSDomainDAO dbs.DAO

func NewNSDomainDAO() *NSDomainDAO {
	return dbs.NewDAO(&NSDomainDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSDomains",
			Model:  new(NSDomain),
			PkName: "id",
		},
	}).(*NSDomainDAO)
}

var SharedNSDomainDAO *NSDomainDAO

func init() {
	dbs.OnReady(func() {
		SharedNSDomainDAO = NewNSDomainDAO()
	})
}

// EnableNSDomain 启用条目
func (this *NSDomainDAO) EnableNSDomain(tx *dbs.Tx, domainId int64) error {
	_, err := this.Query(tx).
		Pk(domainId).
		Set("state", NSDomainStateEnabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, domainId)
}

// DisableNSDomain 禁用条目
func (this *NSDomainDAO) DisableNSDomain(tx *dbs.Tx, domainId int64) error {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(domainId).
		Set("state", NSDomainStateDisabled).
		Set("version", version).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, domainId)
}

// FindEnabledNSDomain 查找启用中的条目
func (this *NSDomainDAO) FindEnabledNSDomain(tx *dbs.Tx, id int64) (*NSDomain, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NSDomainStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NSDomain), err
}

// FindNSDomainName 根据主键查找名称
func (this *NSDomainDAO) FindNSDomainName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateDomain 创建域名
func (this *NSDomainDAO) CreateDomain(tx *dbs.Tx, clusterId int64, userId int64, name string) (int64, error) {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return 0, err
	}

	op := NewNSDomainOperator()
	op.ClusterId = clusterId
	op.UserId = userId
	op.Name = name
	op.Version = version
	op.IsOn = true
	op.State = NSDomainStateEnabled
	domainId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}

	err = this.NotifyUpdate(tx, domainId)
	if err != nil {
		return domainId, err
	}
	return domainId, nil
}

// UpdateDomain 修改域名
func (this *NSDomainDAO) UpdateDomain(tx *dbs.Tx, domainId int64, clusterId int64, userId int64, isOn bool) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}

	oldClusterId, err := this.Query(tx).
		Pk(domainId).
		Result("clusterId").
		FindInt64Col(0)
	if err != nil {
		return err
	}

	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	op := NewNSDomainOperator()
	op.Id = domainId
	op.ClusterId = clusterId
	op.UserId = userId
	op.IsOn = isOn
	op.Version = version
	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	// 通知更新
	if oldClusterId > 0 && oldClusterId != clusterId {
		err = models.SharedNSClusterDAO.NotifyUpdate(tx, oldClusterId)
		if err != nil {
			return err
		}
	}

	return this.NotifyUpdate(tx, domainId)
}

// CountAllEnabledDomains 计算域名数量
func (this *NSDomainDAO) CountAllEnabledDomains(tx *dbs.Tx, clusterId int64, userId int64, keyword string) (int64, error) {
	query := this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	} else {
		query.Where("clusterId IN (SELECT id FROM " + models.SharedNSClusterDAO.Table + " WHERE state=1)")
	}
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		query.Where("(userId=0 OR userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1))")
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}

	return query.
		State(NSDomainStateEnabled).
		Count()
}

// ListEnabledDomains 列出单页域名
func (this *NSDomainDAO) ListEnabledDomains(tx *dbs.Tx, clusterId int64, userId int64, keyword string, offset int64, size int64) (result []*NSDomain, err error) {
	query := this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	} else {
		query.Where("clusterId IN (SELECT id FROM " + models.SharedNSClusterDAO.Table + " WHERE state=1)")
	}
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		query.Where("(userId=0 OR userId IN (SELECT id FROM " + models.SharedUserDAO.Table + " WHERE state=1))")
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	_, err = query.
		State(NSDomainStateEnabled).
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// IncreaseVersion 增加版本
func (this *NSDomainDAO) IncreaseVersion(tx *dbs.Tx) (int64, error) {
	return models.SharedSysLockerDAO.Increase(tx, "NS_DOMAIN_VERSION", 1)
}

// ListDomainsAfterVersion 列出某个版本后的域名
func (this *NSDomainDAO) ListDomainsAfterVersion(tx *dbs.Tx, version int64, size int64) (result []*NSDomain, err error) {
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

// FindDomainIdWithName 根据名称查找域名
func (this *NSDomainDAO) FindDomainIdWithName(tx *dbs.Tx, clusterId int64, name string) (int64, error) {
	return this.Query(tx).
		Attr("clusterId", clusterId).
		Attr("name", name).
		State(NSDomainStateEnabled).
		ResultPk().
		FindInt64Col(0)
}

// FindEnabledDomainTSIG 获取TSIG配置
func (this *NSDomainDAO) FindEnabledDomainTSIG(tx *dbs.Tx, domainId int64) ([]byte, error) {
	tsig, err := this.Query(tx).
		Pk(domainId).
		Result("tsig").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	return []byte(tsig), nil
}

// UpdateDomainTSIG 修改TSIG配置
func (this *NSDomainDAO) UpdateDomainTSIG(tx *dbs.Tx, domainId int64, tsigJSON []byte) error {
	version, err := this.IncreaseVersion(tx)
	if err != nil {
		return err
	}

	err = this.Query(tx).
		Pk(domainId).
		Set("tsig", tsigJSON).
		Set("version", version).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, domainId)
}

// FindEnabledDomainClusterId 获取域名的集群ID
func (this *NSDomainDAO) FindEnabledDomainClusterId(tx *dbs.Tx, domainId int64) (int64, error) {
	return this.Query(tx).
		Pk(domainId).
		State(NSDomainStateEnabled).
		Result("clusterId").
		FindInt64Col(0)
}

// NotifyUpdate 通知更改
func (this *NSDomainDAO) NotifyUpdate(tx *dbs.Tx, domainId int64) error {
	clusterId, err := this.Query(tx).
		Result("clusterId").
		Pk(domainId).
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if clusterId > 0 {
		return models.SharedNodeTaskDAO.CreateClusterTask(tx, nodeconfigs.NodeRoleDNS, clusterId, 0, models.NSNodeTaskTypeDomainChanged)
	}

	return nil
}
