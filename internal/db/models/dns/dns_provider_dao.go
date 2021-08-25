package dns

import (
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"time"
)

const (
	DNSProviderStateEnabled  = 1 // 已启用
	DNSProviderStateDisabled = 0 // 已禁用
)

type DNSProviderDAO dbs.DAO

func NewDNSProviderDAO() *DNSProviderDAO {
	return dbs.NewDAO(&DNSProviderDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeDNSProviders",
			Model:  new(DNSProvider),
			PkName: "id",
		},
	}).(*DNSProviderDAO)
}

var SharedDNSProviderDAO *DNSProviderDAO

func init() {
	dbs.OnReady(func() {
		SharedDNSProviderDAO = NewDNSProviderDAO()
	})
}

// EnableDNSProvider 启用条目
func (this *DNSProviderDAO) EnableDNSProvider(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", DNSProviderStateEnabled).
		Update()
	return err
}

// DisableDNSProvider 禁用条目
func (this *DNSProviderDAO) DisableDNSProvider(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", DNSProviderStateDisabled).
		Update()
	return err
}

// FindEnabledDNSProvider 查找启用中的条目
func (this *DNSProviderDAO) FindEnabledDNSProvider(tx *dbs.Tx, id int64) (*DNSProvider, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", DNSProviderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*DNSProvider), err
}

// CreateDNSProvider 创建服务商
func (this *DNSProviderDAO) CreateDNSProvider(tx *dbs.Tx, adminId int64, userId int64, providerType string, name string, apiParamsJSON []byte) (int64, error) {
	op := NewDNSProviderOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.Type = providerType
	op.Name = name
	if len(apiParamsJSON) > 0 {
		op.ApiParams = apiParamsJSON
	}
	op.State = DNSProviderStateEnabled
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateDNSProvider 修改服务商
func (this *DNSProviderDAO) UpdateDNSProvider(tx *dbs.Tx, dnsProviderId int64, name string, apiParamsJSON []byte) error {
	if dnsProviderId <= 0 {
		return errors.New("invalid dnsProviderId")
	}

	op := NewDNSProviderOperator()
	op.Id = dnsProviderId
	op.Name = name

	// 如果留空则表示不修改
	if len(apiParamsJSON) > 0 {
		op.ApiParams = apiParamsJSON
	}

	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return nil
}

// CountAllEnabledDNSProviders 计算服务商数量
func (this *DNSProviderDAO) CountAllEnabledDNSProviders(tx *dbs.Tx, adminId int64, userId int64, keyword string) (int64, error) {
	var query = dbutils.NewQuery(tx, this, adminId, userId)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	return query.State(DNSProviderStateEnabled).
		Count()
}

// ListEnabledDNSProviders 列出单页服务商
func (this *DNSProviderDAO) ListEnabledDNSProviders(tx *dbs.Tx, adminId int64, userId int64, keyword string, offset int64, size int64) (result []*DNSProvider, err error) {
	var query = dbutils.NewQuery(tx, this, adminId, userId)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	_, err = query.
		State(DNSProviderStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledDNSProviders 列出所有服务商
func (this *DNSProviderDAO) FindAllEnabledDNSProviders(tx *dbs.Tx, adminId int64, userId int64) (result []*DNSProvider, err error) {
	_, err = dbutils.NewQuery(tx, this, adminId, userId).
		State(DNSProviderStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledDNSProvidersWithType 查询某个类型下的所有服务商
func (this *DNSProviderDAO) FindAllEnabledDNSProvidersWithType(tx *dbs.Tx, providerType string) (result []*DNSProvider, err error) {
	_, err = this.Query(tx).
		State(DNSProviderStateEnabled).
		Attr("type", providerType).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// UpdateProviderDataUpdatedTime 更新数据更新时间
func (this *DNSProviderDAO) UpdateProviderDataUpdatedTime(tx *dbs.Tx, providerId int64) error {
	_, err := this.Query(tx).
		Pk(providerId).
		Set("dataUpdatedAt", time.Now().Unix()).
		Update()
	return err
}
