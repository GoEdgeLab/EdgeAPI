package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	DNSDomainStateEnabled  = 1 // 已启用
	DNSDomainStateDisabled = 0 // 已禁用
)

type DNSDomainDAO dbs.DAO

func NewDNSDomainDAO() *DNSDomainDAO {
	return dbs.NewDAO(&DNSDomainDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeDNSDomains",
			Model:  new(DNSDomain),
			PkName: "id",
		},
	}).(*DNSDomainDAO)
}

var SharedDNSDomainDAO *DNSDomainDAO

func init() {
	dbs.OnReady(func() {
		SharedDNSDomainDAO = NewDNSDomainDAO()
	})
}

// 启用条目
func (this *DNSDomainDAO) EnableDNSDomain(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", DNSDomainStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *DNSDomainDAO) DisableDNSDomain(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", DNSDomainStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *DNSDomainDAO) FindEnabledDNSDomain(id int64) (*DNSDomain, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", DNSDomainStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*DNSDomain), err
}

// 根据主键查找名称
func (this *DNSDomainDAO) FindDNSDomainName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建域名
func (this *DNSDomainDAO) CreateDomain(providerId int64, name string) (int64, error) {
	op := NewDNSDomainOperator()
	op.ProviderId = providerId
	op.Name = name
	op.State = DNSDomainStateEnabled
	op.IsOn = true
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改域名
func (this *DNSDomainDAO) UpdateDomain(domainId int64, name string, isOn bool) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Name = name
	op.IsOn = isOn
	_, err := this.Save(op)
	if err != nil {
		return err
	}
	return nil
}

// 查询一个服务商下面的所有域名
func (this *DNSDomainDAO) FindAllEnabledDomainsWithProviderId(providerId int64) (result []*DNSDomain, err error) {
	_, err = this.Query().
		State(DNSDomainStateEnabled).
		Attr("providerId", providerId).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 计算某个服务商下的域名数量
func (this *DNSDomainDAO) CountAllEnabledDomainsWithProviderId(providerId int64) (int64, error) {
	return this.Query().
		State(DNSDomainStateEnabled).
		Attr("providerId", providerId).
		Count()
}

// 更新域名数据
func (this *DNSDomainDAO) UpdateDomainData(domainId int64, data string) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Data = data
	_, err := this.Save(op)
	return err
}

// 更新服务相关域名
func (this *DNSDomainDAO) UpdateServerDomains(domainId int64, serverDomainsJSON []byte) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.ServerDomains = serverDomainsJSON
	_, err := this.Save(op)
	return err
}

// 更新集群相关域名
func (this *DNSDomainDAO) UpdateClusterDomains(domainId int64, clusterDomainJSON []byte) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.ClusterDomains = clusterDomainJSON
	_, err := this.Save(op)
	return err
}

// 更新线路
func (this *DNSDomainDAO) UpdateRoutes(domainId int64, routesJSON []byte) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Routes = routesJSON
	_, err := this.Save(op)
	return err
}
