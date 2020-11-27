package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

const (
	SSLCertStateEnabled  = 1 // 已启用
	SSLCertStateDisabled = 0 // 已禁用
)

type SSLCertDAO dbs.DAO

func NewSSLCertDAO() *SSLCertDAO {
	return dbs.NewDAO(&SSLCertDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeSSLCerts",
			Model:  new(SSLCert),
			PkName: "id",
		},
	}).(*SSLCertDAO)
}

var SharedSSLCertDAO *SSLCertDAO

func init() {
	dbs.OnReady(func() {
		SharedSSLCertDAO = NewSSLCertDAO()
	})
}

// 初始化
func (this *SSLCertDAO) Init() {
	this.DAOObject.Init()
	this.DAOObject.OnUpdate(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
	this.DAOObject.OnInsert(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
	this.DAOObject.OnDelete(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
}

// 启用条目
func (this *SSLCertDAO) EnableSSLCert(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", SSLCertStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *SSLCertDAO) DisableSSLCert(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", SSLCertStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *SSLCertDAO) FindEnabledSSLCert(id int64) (*SSLCert, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", SSLCertStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*SSLCert), err
}

// 根据主键查找名称
func (this *SSLCertDAO) FindSSLCertName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建证书
func (this *SSLCertDAO) CreateCert(adminId int64, userId int64, isOn bool, name string, description string, serverName string, isCA bool, certData []byte, keyData []byte, timeBeginAt int64, timeEndAt int64, dnsNames []string, commonNames []string) (int64, error) {
	op := NewSSLCertOperator()
	op.AdminId = adminId
	op.UserId = userId
	op.State = SSLCertStateEnabled
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	op.ServerName = serverName
	op.IsCA = isCA
	op.CertData = certData
	op.KeyData = keyData
	op.TimeBeginAt = timeBeginAt
	op.TimeEndAt = timeEndAt

	dnsNamesJSON, err := json.Marshal(dnsNames)
	if err != nil {
		return 0, err
	}
	op.DnsNames = dnsNamesJSON

	commonNamesJSON, err := json.Marshal(commonNames)
	if err != nil {
		return 0, err
	}
	op.CommonNames = commonNamesJSON

	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改证书
func (this *SSLCertDAO) UpdateCert(certId int64, isOn bool, name string, description string, serverName string, isCA bool, certData []byte, keyData []byte, timeBeginAt int64, timeEndAt int64, dnsNames []string, commonNames []string) error {
	if certId <= 0 {
		return errors.New("invalid certId")
	}
	op := NewSSLCertOperator()
	op.Id = certId
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	op.ServerName = serverName
	op.IsCA = isCA

	// cert和key均为有重新上传才会修改
	if len(certData) > 0 {
		op.CertData = certData
	}
	if len(keyData) > 0 {
		op.KeyData = keyData
	}

	op.TimeBeginAt = timeBeginAt
	op.TimeEndAt = timeEndAt

	dnsNamesJSON, err := json.Marshal(dnsNames)
	if err != nil {
		return err
	}
	op.DnsNames = dnsNamesJSON

	commonNamesJSON, err := json.Marshal(commonNames)
	if err != nil {
		return err
	}
	op.CommonNames = commonNamesJSON

	_, err = this.Save(op)
	return err
}

// 组合配置
func (this *SSLCertDAO) ComposeCertConfig(certId int64) (*sslconfigs.SSLCertConfig, error) {
	cert, err := this.FindEnabledSSLCert(certId)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, nil
	}

	config := &sslconfigs.SSLCertConfig{}
	config.Id = int64(cert.Id)
	config.IsOn = cert.IsOn == 1
	config.IsCA = cert.IsCA == 1
	config.IsACME = cert.IsACME == 1
	config.Name = cert.Name
	config.Description = cert.Description
	config.CertData = []byte(cert.CertData)
	config.KeyData = []byte(cert.KeyData)
	config.ServerName = cert.ServerName
	config.TimeBeginAt = int64(cert.TimeBeginAt)
	config.TimeEndAt = int64(cert.TimeEndAt)

	if IsNotNull(cert.DnsNames) {
		dnsNames := []string{}
		err := json.Unmarshal([]byte(cert.DnsNames), &dnsNames)
		if err != nil {
			return nil, err
		}
		config.DNSNames = dnsNames
	}

	if IsNotNull(cert.CommonNames) {
		commonNames := []string{}
		err := json.Unmarshal([]byte(cert.CommonNames), &commonNames)
		if err != nil {
			return nil, err
		}
		config.CommonNames = commonNames
	}

	return config, nil
}

// 计算符合条件的证书数量
func (this *SSLCertDAO) CountCerts(isCA bool, isAvailable bool, isExpired bool, expiringDays int64, keyword string) (int64, error) {
	query := this.Query().
		State(SSLCertStateEnabled)
	if isCA {
		query.Attr("isCA", true)
	}
	if isAvailable {
		query.Where("timeBeginAt<=UNIX_TIMESTAMP() AND timeEndAt>=UNIX_TIMESTAMP()")
	}
	if isExpired {
		query.Where("timeEndAt<UNIX_TIMESTAMP()")
	}
	if expiringDays > 0 {
		query.Where("timeEndAt>UNIX_TIMESTAMP() AND timeEndAt<:expiredAt").
			Param("expiredAt", time.Now().Unix()+expiringDays*86400)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword OR dnsNames LIKE :keyword OR commonNames LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}
	return query.Count()
}

// 列出符合条件的证书
func (this *SSLCertDAO) ListCertIds(isCA bool, isAvailable bool, isExpired bool, expiringDays int64, keyword string, offset int64, size int64) (certIds []int64, err error) {
	query := this.Query().
		State(SSLCertStateEnabled)
	if isCA {
		query.Attr("isCA", true)
	}
	if isAvailable {
		query.Where("timeBeginAt<=UNIX_TIMESTAMP() AND timeEndAt>=UNIX_TIMESTAMP()")
	}
	if isExpired {
		query.Where("timeEndAt<UNIX_TIMESTAMP()")
	}
	if expiringDays > 0 {
		query.Where("timeEndAt>UNIX_TIMESTAMP() AND timeEndAt<:expiredAt").
			Param("expiredAt", time.Now().Unix()+expiringDays*86400)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword OR dnsNames LIKE :keyword OR commonNames LIKE :keyword)").
			Param("keyword", "%"+keyword+"%")
	}

	ones, err := query.
		ResultPk().
		DescPk().
		Offset(offset).
		Limit(size).
		FindAll()
	if err != nil {
		return nil, err
	}

	result := []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*SSLCert).Id))
	}
	return result, nil
}

// 设置证书的ACME信息
func (this *SSLCertDAO) UpdateCertACME(certId int64, acmeTaskId int64) error {
	if certId <= 0 {
		return errors.New("invalid certId")
	}
	op := NewSSLCertOperator()
	op.Id = certId
	op.AcmeTaskId = acmeTaskId
	op.IsACME = true
	_, err := this.Save(op)
	return err
}

// 查找需要自动更新的任务
// 这里我们只返回有限的字段以节省内存
func (this *SSLCertDAO) FindAllExpiringCerts(days int) (result []*SSLCert, err error) {
	if days < 0 {
		days = 0
	}

	deltaSeconds := int64(days * 86400)
	_, err = this.Query().
		State(SSLCertStateEnabled).
		Where("FROM_UNIXTIME(timeEndAt, '%Y-%m-%d')=:day AND FROM_UNIXTIME(notifiedAt, '%Y-%m-%d')!=:today").
		Param("day", timeutil.FormatTime("Y-m-d", time.Now().Unix()+deltaSeconds)).
		Param("today", timeutil.Format("Y-m-d")).
		Result("id", "adminId", "userId", "timeEndAt", "name", "dnsNames", "notifiedAt", "acmeTaskId").
		Slice(&result).
		AscPk().
		FindAll()
	return
}

// 设置当前证书事件通知时间
func (this *SSLCertDAO) UpdateCertNotifiedAt(certId int64) error {
	_, err := this.Query().
		Pk(certId).
		Set("notifiedAt", time.Now().Unix()).
		Update()
	return err
}
