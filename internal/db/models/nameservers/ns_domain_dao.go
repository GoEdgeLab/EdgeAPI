package nameservers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
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
func (this *NSDomainDAO) EnableNSDomain(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NSDomainStateEnabled).
		Update()
	return err
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
	return err
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
	return this.SaveInt64(tx, op)
}

// UpdateDomain 修改域名
func (this *NSDomainDAO) UpdateDomain(tx *dbs.Tx, domainId int64, clusterId int64, userId int64, isOn bool) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
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
	return this.Save(tx, op)
}

// CountAllEnabledDomains 计算域名数量
func (this *NSDomainDAO) CountAllEnabledDomains(tx *dbs.Tx, clusterId int64, userId int64, keyword string) (int64, error) {
	query := this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	} else {
		query.Where("clusterId IN (SELECT id FROM " + SharedNSClusterDAO.Table + " WHERE state=1)")
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
		query.Where("clusterId IN (SELECT id FROM " + SharedNSClusterDAO.Table + " WHERE state=1)")
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
