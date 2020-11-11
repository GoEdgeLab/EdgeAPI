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

// 启用条目
func (this *DNSProviderDAO) EnableDNSProvider(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", DNSProviderStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *DNSProviderDAO) DisableDNSProvider(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", DNSProviderStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *DNSProviderDAO) FindEnabledDNSProvider(id int64) (*DNSProvider, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", DNSProviderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*DNSProvider), err
}

// 创建服务商
func (this *DNSProviderDAO) CreateDNSProvider(providerType string, name string, apiParamsJSON []byte, routesJSON []byte) (int64, error) {
	op := NewDNSProviderOperator()
	op.Type = providerType
	op.Name = name
	if len(apiParamsJSON) > 0 {
		op.ApiParams = apiParamsJSON
	}
	if len(routesJSON) > 0 {
		op.Routes = routesJSON
	}
	op.State = DNSProviderStateEnabled
	op.DataUpdatedAt = time.Now().Unix()
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改服务商
func (this *DNSProviderDAO) UpdateDNSProvider(dnsProviderId int64, name string, apiParamsJSON []byte, routesJSON []byte) error {
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

	// 如果留空则表示不修改
	if len(routesJSON) > 0 {
		op.Routes = routesJSON
	}

	op.DataUpdatedAt = time.Now().Unix()
	_, err := this.Save(op)
	if err != nil {
		return err
	}
	return nil
}

// 计算服务商数量
func (this *DNSProviderDAO) CountAllEnabledDNSProviders() (int64, error) {
	return this.Query().
		State(DNSProviderStateEnabled).
		Count()
}

// 列出单页服务商
func (this *DNSProviderDAO) ListEnabledDNSProviders(offset int64, size int64) (result []*DNSProvider, err error) {
	_, err = this.Query().
		State(DNSProviderStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
