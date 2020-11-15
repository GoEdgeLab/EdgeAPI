package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"time"
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

// 更新域名解析记录
func (this *DNSDomainDAO) UpdateDomainRecords(domainId int64, recordsJSON []byte) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Records = recordsJSON
	op.DataUpdatedAt = time.Now().Unix()
	_, err := this.Save(op)
	return err
}

// 更新线路
func (this *DNSDomainDAO) UpdateDomainRoutes(domainId int64, routesJSON []byte) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Routes = routesJSON
	op.DataUpdatedAt = time.Now().Unix()
	_, err := this.Save(op)
	return err
}

// 查找域名线路
func (this *DNSDomainDAO) FindDomainRoutes(domainId int64) ([]*dnsclients.Route, error) {
	routes, err := this.Query().
		Pk(domainId).
		Result("routes").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(routes) == 0 || routes == "null" {
		return nil, nil
	}
	result := []*dnsclients.Route{}
	err = json.Unmarshal([]byte(routes), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 查找线路名称
func (this *DNSDomainDAO) FindDomainRouteName(domainId int64, routeCode string) (string, error) {
	routes, err := this.FindDomainRoutes(domainId)
	if err != nil {
		return "", err
	}
	for _, route := range routes {
		if route.Code == routeCode {
			return route.Name, nil
		}
	}
	return "", nil
}

// 判断是否有域名可选
func (this *DNSDomainDAO) ExistAvailableDomains() (bool, error) {
	subQuery, err := SharedDNSProviderDAO.Query().
		Where("state=1"). // 这里要使用非变量
		ResultPk().
		AsSQL()
	if err != nil {
		return false, err
	}
	return this.Query().
		State(DNSDomainStateEnabled).
		Attr("isOn", true).
		Where("providerId IN (" + subQuery + ")").
		Exist()
}
