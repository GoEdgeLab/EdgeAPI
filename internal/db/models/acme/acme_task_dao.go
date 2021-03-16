package acme

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
	"github.com/go-acme/lego/v4/registration"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"time"
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
func (this *ACMETaskDAO) EnableACMETask(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ACMETaskStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *ACMETaskDAO) DisableACMETask(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ACMETaskStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *ACMETaskDAO) FindEnabledACMETask(tx *dbs.Tx, id int64) (*ACMETask, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", ACMETaskStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ACMETask), err
}

// 计算某个ACME用户相关的任务数量
func (this *ACMETaskDAO) CountACMETasksWithACMEUserId(tx *dbs.Tx, acmeUserId int64) (int64, error) {
	return this.Query(tx).
		State(ACMETaskStateEnabled).
		Attr("acmeUserId", acmeUserId).
		Count()
}

// 计算某个DNS服务商相关的任务数量
func (this *ACMETaskDAO) CountACMETasksWithDNSProviderId(tx *dbs.Tx, dnsProviderId int64) (int64, error) {
	return this.Query(tx).
		State(ACMETaskStateEnabled).
		Attr("dnsProviderId", dnsProviderId).
		Count()
}

// 停止某个证书相关任务
func (this *ACMETaskDAO) DisableAllTasksWithCertId(tx *dbs.Tx, certId int64) error {
	_, err := this.Query(tx).
		Attr("certId", certId).
		Set("state", ACMETaskStateDisabled).
		Update()
	return err
}

// 计算所有任务数量
func (this *ACMETaskDAO) CountAllEnabledACMETasks(tx *dbs.Tx, adminId int64, userId int64, isAvailable bool, isExpired bool, expiringDays int64, keyword string) (int64, error) {
	query := dbutils.NewQuery(tx, this, adminId, userId)
	if isAvailable || isExpired || expiringDays > 0 {
		query.Gt("certId", 0)

		if isAvailable {
			query.Where("certId IN (SELECT id FROM " + models.SharedSSLCertDAO.Table + " WHERE timeBeginAt<=UNIX_TIMESTAMP() AND timeEndAt>=UNIX_TIMESTAMP())")
		}
		if isExpired {
			query.Where("certId IN (SELECT id FROM " + models.SharedSSLCertDAO.Table + " WHERE timeEndAt<UNIX_TIMESTAMP())")
		}
		if expiringDays > 0 {
			query.Where("certId IN (SELECT id FROM "+models.SharedSSLCertDAO.Table+" WHERE timeEndAt>UNIX_TIMESTAMP() AND timeEndAt<:expiredAt)").
				Param("expiredAt", time.Now().Unix()+expiringDays*86400)
		}
	}

	if len(keyword) > 0 {
		query.Where("(domains LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	if len(keyword) > 0 {
		query.Where("domains LIKE :keyword").
			Param("keyword", "%"+keyword+"%")
	}

	return query.State(ACMETaskStateEnabled).
		Count()
}

// 列出单页任务
func (this *ACMETaskDAO) ListEnabledACMETasks(tx *dbs.Tx, adminId int64, userId int64, isAvailable bool, isExpired bool, expiringDays int64, keyword string, offset int64, size int64) (result []*ACMETask, err error) {
	query := dbutils.NewQuery(tx, this, adminId, userId)
	if isAvailable || isExpired || expiringDays > 0 {
		query.Gt("certId", 0)

		if isAvailable {
			query.Where("certId IN (SELECT id FROM " + models.SharedSSLCertDAO.Table + " WHERE timeBeginAt<=UNIX_TIMESTAMP() AND timeEndAt>=UNIX_TIMESTAMP())")
		}
		if isExpired {
			query.Where("certId IN (SELECT id FROM " + models.SharedSSLCertDAO.Table + " WHERE timeEndAt<UNIX_TIMESTAMP())")
		}
		if expiringDays > 0 {
			query.Where("certId IN (SELECT id FROM "+models.SharedSSLCertDAO.Table+" WHERE timeEndAt>UNIX_TIMESTAMP() AND timeEndAt<:expiredAt)").
				Param("expiredAt", time.Now().Unix()+expiringDays*86400)
		}
	}
	if len(keyword) > 0 {
		query.Where("(domains LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	_, err = query.
		State(ACMETaskStateEnabled).
		DescPk().
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// 创建任务
func (this *ACMETaskDAO) CreateACMETask(tx *dbs.Tx, adminId int64, userId int64, authType acme.AuthType, acmeUserId int64, dnsProviderId int64, dnsDomain string, domains []string, autoRenew bool) (int64, error) {
	op := NewACMETaskOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.AuthType = authType
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
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改任务
func (this *ACMETaskDAO) UpdateACMETask(tx *dbs.Tx, acmeTaskId int64, acmeUserId int64, dnsProviderId int64, dnsDomain string, domains []string, autoRenew bool) error {
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
	err := this.Save(tx, op)
	return err
}

// 检查权限
func (this *ACMETaskDAO) CheckACMETask(tx *dbs.Tx, adminId int64, userId int64, acmeTaskId int64) (bool, error) {
	return dbutils.NewQuery(tx, this, adminId, userId).
		State(ACMETaskStateEnabled).
		Pk(acmeTaskId).
		Exist()
}

// 设置任务关联的证书
func (this *ACMETaskDAO) UpdateACMETaskCert(tx *dbs.Tx, taskId int64, certId int64) error {
	if taskId <= 0 {
		return errors.New("invalid taskId")
	}

	op := NewACMETaskOperator()
	op.Id = taskId
	op.CertId = certId
	err := this.Save(tx, op)
	return err
}

// 执行任务并记录日志
func (this *ACMETaskDAO) RunTask(tx *dbs.Tx, taskId int64) (isOk bool, errMsg string, resultCertId int64) {
	isOk, errMsg, resultCertId = this.runTaskWithoutLog(tx, taskId)

	// 记录日志
	err := SharedACMETaskLogDAO.CreateACMETaskLog(tx, taskId, isOk, errMsg)
	if err != nil {
		logs.Error(err)
	}

	return
}

// 执行任务但并不记录日志
func (this *ACMETaskDAO) runTaskWithoutLog(tx *dbs.Tx, taskId int64) (isOk bool, errMsg string, resultCertId int64) {
	task, err := this.FindEnabledACMETask(tx, taskId)
	if err != nil {
		errMsg = "查询任务信息时出错：" + err.Error()
		return
	}
	if task == nil {
		errMsg = "找不到要执行的任务"
		return
	}
	if task.IsOn != 1 {
		errMsg = "任务没有启用"
		return
	}

	// ACME用户
	user, err := SharedACMEUserDAO.FindEnabledACMEUser(tx, int64(task.AcmeUserId))
	if err != nil {
		errMsg = "查询ACME用户时出错：" + err.Error()
		return
	}
	if user == nil {
		errMsg = "找不到ACME用户"
		return
	}

	privateKey, err := acme.ParsePrivateKeyFromBase64(user.PrivateKey)
	if err != nil {
		errMsg = "解析私钥时出错：" + err.Error()
		return
	}

	remoteUser := acme.NewUser(user.Email, privateKey, func(resource *registration.Resource) error {
		resourceJSON, err := json.Marshal(resource)
		if err != nil {
			return err
		}

		err = SharedACMEUserDAO.UpdateACMEUserRegistration(tx, int64(user.Id), resourceJSON)
		return err
	})

	if len(user.Registration) > 0 {
		err = remoteUser.SetRegistration([]byte(user.Registration))
		if err != nil {
			errMsg = "设置注册信息时出错：" + err.Error()
			return
		}
	}

	var acmeTask *acme.Task = nil
	if task.AuthType == acme.AuthTypeDNS {
		// DNS服务商
		dnsProvider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(task.DnsProviderId))
		if err != nil {
			errMsg = "查找DNS服务商账号信息时出错：" + err.Error()
			return
		}
		if dnsProvider == nil {
			errMsg = "找不到DNS服务商账号"
			return
		}
		providerInterface := dnsclients.FindProvider(dnsProvider.Type)
		if providerInterface == nil {
			errMsg = "暂不支持此类型的DNS服务商 '" + dnsProvider.Type + "'"
			return
		}
		apiParams, err := dnsProvider.DecodeAPIParams()
		if err != nil {
			errMsg = "解析DNS服务商API参数时出错：" + err.Error()
			return
		}
		err = providerInterface.Auth(apiParams)
		if err != nil {
			errMsg = "校验DNS服务商API参数时出错：" + err.Error()
			return
		}

		acmeTask = &acme.Task{
			User:        remoteUser,
			AuthType:    acme.AuthTypeDNS,
			DNSProvider: providerInterface,
			DNSDomain:   task.DnsDomain,
			Domains:     task.DecodeDomains(),
		}
	} else if task.AuthType == acme.AuthTypeHTTP {
		acmeTask = &acme.Task{
			User:     remoteUser,
			AuthType: acme.AuthTypeHTTP,
			Domains:  task.DecodeDomains(),
		}
	}

	acmeRequest := acme.NewRequest(acmeTask)
	acmeRequest.OnAuth(func(domain, token, keyAuth string) {
		err := SharedACMEAuthenticationDAO.CreateAuth(tx, taskId, domain, token, keyAuth)
		if err != nil {
			logs.Println("[ACME]write authentication to database error: " + err.Error())
		}
	})
	certData, keyData, err := acmeRequest.Run()
	if err != nil {
		errMsg = "证书生成失败：" + err.Error()
		return
	}

	// 分析证书
	sslConfig := &sslconfigs.SSLCertConfig{
		CertData: certData,
		KeyData:  keyData,
	}
	err = sslConfig.Init()
	if err != nil {
		errMsg = "证书生成成功，但是分析证书信息时发生错误：" + err.Error()
		return
	}

	// 保存证书
	resultCertId = int64(task.CertId)
	if resultCertId > 0 {
		cert, err := models.SharedSSLCertDAO.FindEnabledSSLCert(tx, resultCertId)
		if err != nil {
			errMsg = "证书生成成功，但查询已绑定的证书时出错：" + err.Error()
			return
		}
		if cert == nil {
			errMsg = "证书已被管理员或用户删除"

			// 禁用
			err = SharedACMETaskDAO.DisableACMETask(tx, taskId)
			if err != nil {
				errMsg = "禁用失效的ACME任务出错：" + err.Error()
			}

			return
		}

		err = models.SharedSSLCertDAO.UpdateCert(tx, resultCertId, cert.IsOn == 1, cert.Name, cert.Description, cert.ServerName, cert.IsCA == 1, certData, keyData, sslConfig.TimeBeginAt, sslConfig.TimeEndAt, sslConfig.DNSNames, sslConfig.CommonNames)
		if err != nil {
			errMsg = "证书生成成功，但是修改数据库中的证书信息时出错：" + err.Error()
			return
		}
	} else {
		resultCertId, err = models.SharedSSLCertDAO.CreateCert(tx, int64(task.AdminId), int64(task.UserId), true, task.DnsDomain+"免费证书", "免费申请的证书", "", false, certData, keyData, sslConfig.TimeBeginAt, sslConfig.TimeEndAt, sslConfig.DNSNames, sslConfig.CommonNames)
		if err != nil {
			errMsg = "证书生成成功，但是保存到数据库失败：" + err.Error()
			return
		}

		err = models.SharedSSLCertDAO.UpdateCertACME(tx, resultCertId, int64(task.Id))
		if err != nil {
			errMsg = "证书生成成功，修改证书ACME信息时出错：" + err.Error()
			return
		}

		// 设置成功
		err = SharedACMETaskDAO.UpdateACMETaskCert(tx, taskId, resultCertId)
		if err != nil {
			errMsg = "证书生成成功，设置任务关联的证书时出错：" + err.Error()
			return
		}
	}

	isOk = true
	return
}
