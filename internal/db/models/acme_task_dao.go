package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	ACMETaskStateEnabled  = 1 // 已启用
	ACMETaskStateDisabled = 0 // 已禁用
)

type ACMETaskDAO dbs.DAO

func NewACMETaskDAO() *ACMETaskDAO {
	return dbs.NewDAO(&ACMETaskDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeACMETasks",
			Model:  new(ACMETask),
			PkName: "id",
		},
	}).(*ACMETaskDAO)
}

var SharedACMETaskDAO *ACMETaskDAO

func init() {
	dbs.OnReady(func() {
		SharedACMETaskDAO = NewACMETaskDAO()
	})
}

// 启用条目
func (this *ACMETaskDAO) EnableACMETask(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ACMETaskStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *ACMETaskDAO) DisableACMETask(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", ACMETaskStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *ACMETaskDAO) FindEnabledACMETask(id int64) (*ACMETask, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", ACMETaskStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ACMETask), err
}

// 计算某个ACME用户相关的任务数量
func (this *ACMETaskDAO) CountACMETasksWithACMEUserId(acmeUserId int64) (int64, error) {
	return this.Query().
		State(ACMETaskStateEnabled).
		Attr("acmeUserId", acmeUserId).
		Count()
}

// 计算某个DNS服务商相关的任务数量
func (this *ACMETaskDAO) CountACMETasksWithDNSProviderId(dnsProviderId int64) (int64, error) {
	return this.Query().
		State(ACMETaskStateEnabled).
		Attr("dnsProviderId", dnsProviderId).
		Count()
}

// 停止某个证书相关任务
func (this *ACMETaskDAO) DisableAllTasksWithCertId(certId int64) error {
	_, err := this.Query().
		Attr("certId", certId).
		Set("state", ACMETaskStateDisabled).
		Update()
	return err
}

// 计算所有任务数量
func (this *ACMETaskDAO) CountAllEnabledACMETasks(adminId int64, userId int64) (int64, error) {
	return NewQuery(this, adminId, userId).
		State(ACMETaskStateEnabled).
		Count()
}

// 列出单页任务
func (this *ACMETaskDAO) ListEnabledACMETasks(adminId int64, userId int64, offset int64, size int64) (result []*ACMETask, err error) {
	_, err = NewQuery(this, adminId, userId).
		State(ACMETaskStateEnabled).
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// 创建任务
func (this *ACMETaskDAO) CreateACMETask(adminId int64, userId int64, acmeUserId int64, dnsProviderId int64, dnsDomain string, domains []string, autoRenew bool) (int64, error) {
	op := NewACMETaskOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.AcmeUserId = acmeUserId
	op.DnsProviderId = dnsProviderId
	op.DnsDomain = dnsDomain

	if len(domains) > 0 {
		domainsJSON, err := json.Marshal(domains)
		if err != nil {
			return 0, err
		}
		op.Domains = domainsJSON
	} else {
		op.Domains = "[]"
	}

	op.AutoRenew = autoRenew
	op.IsOn = true
	op.State = ACMETaskStateEnabled
	op.IsOk = false
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改任务
func (this *ACMETaskDAO) UpdateACMETask(acmeTaskId int64, acmeUserId int64, dnsProviderId int64, dnsDomain string, domains []string, autoRenew bool) error {
	if acmeTaskId <= 0 {
		return errors.New("invalid acmeTaskId")
	}

	op := NewACMETaskOperator()
	op.Id = acmeTaskId
	op.AcmeUserId = acmeUserId
	op.DnsProviderId = dnsProviderId
	op.DnsDomain = dnsDomain

	if len(domains) > 0 {
		domainsJSON, err := json.Marshal(domains)
		if err != nil {
			return err
		}
		op.Domains = domainsJSON
	} else {
		op.Domains = "[]"
	}

	op.AutoRenew = autoRenew
	_, err := this.Save(op)
	return err
}

// 检查权限
func (this *ACMETaskDAO) CheckACMETask(adminId int64, userId int64, acmeTaskId int64) (bool, error) {
	return NewQuery(this, adminId, userId).
		State(ACMETaskStateEnabled).
		Pk(acmeTaskId).
		Exist()
}
