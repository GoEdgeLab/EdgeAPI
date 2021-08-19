package dns

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"strings"
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

// EnableDNSDomain 启用条目
func (this *DNSDomainDAO) EnableDNSDomain(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", DNSDomainStateEnabled).
		Update()
	return err
}

// DisableDNSDomain 禁用条目
func (this *DNSDomainDAO) DisableDNSDomain(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", DNSDomainStateDisabled).
		Update()
	return err
}

// FindEnabledDNSDomain 查找启用中的条目
func (this *DNSDomainDAO) FindEnabledDNSDomain(tx *dbs.Tx, id int64) (*DNSDomain, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", DNSDomainStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*DNSDomain), err
}

// FindDNSDomainName 根据主键查找名称
func (this *DNSDomainDAO) FindDNSDomainName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateDomain 创建域名
func (this *DNSDomainDAO) CreateDomain(tx *dbs.Tx, adminId int64, userId int64, providerId int64, name string) (int64, error) {
	op := NewDNSDomainOperator()
	op.ProviderId = providerId
	op.AdminId = adminId
	op.UserId = userId
	op.Name = name
	op.State = DNSDomainStateEnabled
	op.IsOn = true
	op.IsUp = true
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateDomain 修改域名
func (this *DNSDomainDAO) UpdateDomain(tx *dbs.Tx, domainId int64, name string, isOn bool) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Name = name
	op.IsOn = isOn
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return nil
}

// FindAllEnabledDomainsWithProviderId 查询一个服务商下面的所有域名
func (this *DNSDomainDAO) FindAllEnabledDomainsWithProviderId(tx *dbs.Tx, providerId int64) (result []*DNSDomain, err error) {
	_, err = this.Query(tx).
		State(DNSDomainStateEnabled).
		Attr("providerId", providerId).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledDomainsWithProviderId 计算某个服务商下的域名数量
func (this *DNSDomainDAO) CountAllEnabledDomainsWithProviderId(tx *dbs.Tx, providerId int64) (int64, error) {
	return this.Query(tx).
		State(DNSDomainStateEnabled).
		Attr("providerId", providerId).
		Count()
}

// UpdateDomainData 更新域名数据
func (this *DNSDomainDAO) UpdateDomainData(tx *dbs.Tx, domainId int64, data string) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Data = data
	err := this.Save(tx, op)
	return err
}

// UpdateDomainRecords 更新域名解析记录
func (this *DNSDomainDAO) UpdateDomainRecords(tx *dbs.Tx, domainId int64, recordsJSON []byte) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Records = recordsJSON
	op.DataUpdatedAt = time.Now().Unix()
	err := this.Save(tx, op)
	return err
}

// UpdateDomainRoutes 更新线路
func (this *DNSDomainDAO) UpdateDomainRoutes(tx *dbs.Tx, domainId int64, routesJSON []byte) error {
	if domainId <= 0 {
		return errors.New("invalid domainId")
	}
	op := NewDNSDomainOperator()
	op.Id = domainId
	op.Routes = routesJSON
	op.DataUpdatedAt = time.Now().Unix()
	err := this.Save(tx, op)
	return err
}

// FindDomainRoutes 查找域名线路
func (this *DNSDomainDAO) FindDomainRoutes(tx *dbs.Tx, domainId int64) ([]*dnstypes.Route, error) {
	routes, err := this.Query(tx).
		Pk(domainId).
		Result("routes").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if len(routes) == 0 || routes == "null" {
		return nil, nil
	}
	result := []*dnstypes.Route{}
	err = json.Unmarshal([]byte(routes), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindDomainRouteName 查找线路名称
func (this *DNSDomainDAO) FindDomainRouteName(tx *dbs.Tx, domainId int64, routeCode string) (string, error) {
	routes, err := this.FindDomainRoutes(tx, domainId)
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

// ExistAvailableDomains 判断是否有域名可选
func (this *DNSDomainDAO) ExistAvailableDomains(tx *dbs.Tx) (bool, error) {
	subQuery, err := SharedDNSProviderDAO.Query(tx).
		Where("state=1"). // 这里要使用非变量
		ResultPk().
		AsSQL()
	if err != nil {
		return false, err
	}
	return this.Query(tx).
		State(DNSDomainStateEnabled).
		Attr("isOn", true).
		Where("providerId IN (" + subQuery + ")").
		Exist()
}

// ExistDomainRecord 检查域名解析记录是否存在
func (this *DNSDomainDAO) ExistDomainRecord(tx *dbs.Tx, domainId int64, recordName string, recordType string, recordRoute string, recordValue string) (bool, error) {
	recordType = strings.ToUpper(recordType)

	query := maps.Map{
		"name": recordName,
		"type": recordType,
	}
	if len(recordRoute) > 0 {
		query["route"] = recordRoute
	}
	if len(recordValue) > 0 {
		query["value"] = recordValue

		// CNAME兼容点（.）符号
		if recordType == "CNAME" && !strings.HasSuffix(recordValue, ".") {
			b, err := this.ExistDomainRecord(tx, domainId, recordName, recordType, recordRoute, recordValue+".")
			if err != nil {
				return false, err
			}
			if b {
				return true, nil
			}
		}
	}
	return this.Query(tx).
		Pk(domainId).
		Where("JSON_CONTAINS(records, :query)").
		Param("query", query.AsJSON()).
		Exist()
}

// FindEnabledDomainWithName 根据名称查找某个域名
func (this *DNSDomainDAO) FindEnabledDomainWithName(tx *dbs.Tx, providerId int64, domainName string) (*DNSDomain, error) {
	one, err := this.Query(tx).
		State(DNSDomainStateEnabled).
		Attr("isOn", true).
		Attr("providerId", providerId).
		Attr("name", domainName).
		Find()
	if one != nil {
		return one.(*DNSDomain), nil
	}
	return nil, err
}

// UpdateDomainIsUp 设置是否在线
func (this *DNSDomainDAO) UpdateDomainIsUp(tx *dbs.Tx, domainId int64, isUp bool) error {
	return this.Query(tx).
		Pk(domainId).
		Set("isUp", isUp).
		UpdateQuickly()
}
