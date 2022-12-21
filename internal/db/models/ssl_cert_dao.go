package models

import (
	"bytes"
	"encoding/json"
	"errors"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
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

// Init 初始化
func (this *SSLCertDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableSSLCert 启用条目
func (this *SSLCertDAO) EnableSSLCert(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", SSLCertStateEnabled).
		Update()
	return err
}

// DisableSSLCert 禁用条目
func (this *SSLCertDAO) DisableSSLCert(tx *dbs.Tx, certId int64) error {
	_, err := this.Query(tx).
		Pk(certId).
		Set("state", SSLCertStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, certId)
}

// FindEnabledSSLCert 查找启用中的条目
func (this *SSLCertDAO) FindEnabledSSLCert(tx *dbs.Tx, id int64) (*SSLCert, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", SSLCertStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*SSLCert), err
}

// FindSSLCertName 根据主键查找名称
func (this *SSLCertDAO) FindSSLCertName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateCert 创建证书
func (this *SSLCertDAO) CreateCert(tx *dbs.Tx, adminId int64, userId int64, isOn bool, name string, description string, serverName string, isCA bool, certData []byte, keyData []byte, timeBeginAt int64, timeEndAt int64, dnsNames []string, commonNames []string) (int64, error) {
	var op = NewSSLCertOperator()
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

	op.OcspIsUpdated = false

	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateCert 修改证书
func (this *SSLCertDAO) UpdateCert(tx *dbs.Tx,
	certId int64,
	isOn bool,
	name string,
	description string,
	serverName string,
	isCA bool,
	certData []byte,
	keyData []byte,
	timeBeginAt int64,
	timeEndAt int64,
	dnsNames []string, commonNames []string) error {
	if certId <= 0 {
		return errors.New("invalid certId")
	}

	oldOne, err := this.Query(tx).Find()
	if err != nil {
		return err
	}
	if oldOne == nil {
		return nil
	}
	var oldCert = oldOne.(*SSLCert)
	var dataIsChanged = !bytes.Equal(certData, oldCert.CertData) || !bytes.Equal(keyData, oldCert.KeyData)

	var op = NewSSLCertOperator()
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

	// OCSP
	if dataIsChanged {
		op.OcspIsUpdated = 0
		op.Ocsp = ""
		op.OcspUpdatedAt = 0
		op.OcspError = ""
		op.OcspTries = 0
		op.OcspExpiresAt = 0
	}

	err = this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, certId)
}

// ComposeCertConfig 组合配置
func (this *SSLCertDAO) ComposeCertConfig(tx *dbs.Tx, certId int64, cacheMap *utils.CacheMap) (*sslconfigs.SSLCertConfig, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":ComposeCertConfig:" + types.String(certId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*sslconfigs.SSLCertConfig), nil
	}

	cert, err := this.FindEnabledSSLCert(tx, certId)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, nil
	}

	config := &sslconfigs.SSLCertConfig{}
	config.Id = int64(cert.Id)
	config.IsOn = cert.IsOn
	config.IsCA = cert.IsCA
	config.IsACME = cert.IsACME
	config.Name = cert.Name
	config.Description = cert.Description
	config.CertData = cert.CertData
	config.KeyData = cert.KeyData
	config.ServerName = cert.ServerName
	config.TimeBeginAt = int64(cert.TimeBeginAt)
	config.TimeEndAt = int64(cert.TimeEndAt)

	// OCSP
	if int64(cert.OcspExpiresAt) > time.Now().Unix() {
		config.OCSP = cert.Ocsp
		config.OCSPExpiresAt = int64(cert.OcspExpiresAt)
	}
	config.OCSPError = cert.OcspError

	if IsNotNull(cert.DnsNames) {
		dnsNames := []string{}
		err := json.Unmarshal(cert.DnsNames, &dnsNames)
		if err != nil {
			return nil, err
		}
		config.DNSNames = dnsNames
	}

	if cert.CommonNames.IsNotNull() {
		commonNames := []string{}
		err := json.Unmarshal(cert.CommonNames, &commonNames)
		if err != nil {
			return nil, err
		}
		config.CommonNames = commonNames
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// CountCerts 计算符合条件的证书数量
func (this *SSLCertDAO) CountCerts(tx *dbs.Tx, isCA bool, isAvailable bool, isExpired bool, expiringDays int64, keyword string, userId int64) (int64, error) {
	query := this.Query(tx).
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
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		// 只查询管理员上传的
		query.Attr("userId", 0)
	}
	return query.Count()
}

// ListCertIds 列出符合条件的证书
func (this *SSLCertDAO) ListCertIds(tx *dbs.Tx, isCA bool, isAvailable bool, isExpired bool, expiringDays int64, keyword string, userId int64, offset int64, size int64) (certIds []int64, err error) {
	query := this.Query(tx).
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
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	if userId > 0 {
		query.Attr("userId", userId)
	} else {
		// 只查询管理员上传的
		query.Attr("userId", 0)
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

// UpdateCertACME 设置证书的ACME信息
func (this *SSLCertDAO) UpdateCertACME(tx *dbs.Tx, certId int64, acmeTaskId int64) error {
	if certId <= 0 {
		return errors.New("invalid certId")
	}
	var op = NewSSLCertOperator()
	op.Id = certId
	op.AcmeTaskId = acmeTaskId
	op.IsACME = true
	err := this.Save(tx, op)
	return err
}

// FindAllExpiringCerts 查找需要自动更新的任务
// 这里我们只返回有限的字段以节省内存
func (this *SSLCertDAO) FindAllExpiringCerts(tx *dbs.Tx, days int) (result []*SSLCert, err error) {
	if days < 0 {
		days = 0
	}

	var deltaSeconds = int64(days * 86400)
	_, err = this.Query(tx).
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

// UpdateCertNotifiedAt 设置当前证书事件通知时间
func (this *SSLCertDAO) UpdateCertNotifiedAt(tx *dbs.Tx, certId int64) error {
	_, err := this.Query(tx).
		Pk(certId).
		Set("notifiedAt", time.Now().Unix()).
		Update()
	return err
}

// CheckUserCert 检查用户权限
func (this *SSLCertDAO) CheckUserCert(tx *dbs.Tx, certId int64, userId int64) error {
	if certId <= 0 || userId <= 0 {
		return errors.New("not found")
	}
	ok, err := this.Query(tx).
		Pk(certId).
		Attr("userId", userId).
		State(SSLCertStateEnabled).
		Exist()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("not found")
	}
	return nil
}

// UpdateCertUser 修改证书所属用户
func (this *SSLCertDAO) UpdateCertUser(tx *dbs.Tx, certId int64, userId int64) error {
	if certId <= 0 || userId <= 0 {
		return nil
	}
	return this.Query(tx).
		Pk(certId).
		Set("userId", userId).
		UpdateQuickly()
}

// ListCertsToUpdateOCSP 查找需要更新OCSP的证书
func (this *SSLCertDAO) ListCertsToUpdateOCSP(tx *dbs.Tx, maxTries int, size int64) (result []*SSLCert, err error) {
	var nowTime = time.Now().Unix()
	var query = this.Query(tx).
		State(SSLCertStateEnabled).
		Lt("ocspExpiresAt", nowTime+120). // 提前 N 秒钟准备更新
		Lt("ocspTries", maxTries).
		Lt("timeBeginAt", nowTime).
		Gt("timeEndAt", nowTime)

	// TODO 需要排除没有被server使用的policy，或许可以增加一个字段记录policy最近使用时间

	// 检查函数
	var JSONArrayAggIsEnabled = false
	_, err = this.Object().Instance.Exec("SELECT JSON_ARRAYAGG('1')")
	if err == nil {
		JSONArrayAggIsEnabled = true
	}

	if JSONArrayAggIsEnabled {
		query.Where("JSON_CONTAINS((SELECT JSON_ARRAYAGG(JSON_EXTRACT(certs, '$[*].certId')) FROM edgeSSLPolicies WHERE state=1 AND ocspIsOn=1 AND certs IS NOT NULL), CAST(id AS CHAR))")
	} else {
		query.Where("JSON_CONTAINS((SELECT REPLACE(GROUP_CONCAT(JSON_EXTRACT(certs, '$[*].certId')), '],[', ',') FROM edgeSSLPolicies WHERE state=1 AND ocspIsOn=1 AND certs IS NOT NULL), CAST(id AS CHAR))")
	}

	_, err = query.
		Asc("ocspUpdatedAt").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// ListCertOCSPAfterVersion 列出某个版本后的OCSP
func (this *SSLCertDAO) ListCertOCSPAfterVersion(tx *dbs.Tx, version int64, size int64) (result []*SSLCert, err error) {
	// 不需要判断ocsp是否为空
	_, err = this.Query(tx).
		Result("id", "ocsp", "ocspUpdatedVersion", "ocspExpiresAt").
		State(SSLCertStateEnabled).
		Attr("ocspIsUpdated", 1).
		Gt("ocspUpdatedVersion", version).
		Asc("ocspUpdatedVersion").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindCertOCSPLatestVersion 获取OCSP最新版本
func (this *SSLCertDAO) FindCertOCSPLatestVersion(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		Result("ocspUpdatedVersion").
		Desc("ocspUpdatedVersion").
		Limit(1).
		FindInt64Col(0)
}

// PrepareCertOCSPUpdating 更新OCSP更新时间，以便于准备更新，相当于锁定
func (this *SSLCertDAO) PrepareCertOCSPUpdating(tx *dbs.Tx, certId int64) error {
	return this.Query(tx).
		Pk(certId).
		Set("ocspUpdatedAt", time.Now().Unix()).
		UpdateQuickly()

}

// UpdateCertOCSP 修改OCSP
func (this *SSLCertDAO) UpdateCertOCSP(tx *dbs.Tx, certId int64, ocsp []byte, expiresAt int64, hasErr bool, errString string) error {
	if hasErr && len(errString) == 0 {
		errString = "failed"
	}

	version, err := SharedSysLockerDAO.Increase(tx, "SSL_CERT_OCSP_VERSION", 1)
	if err != nil {
		return err
	}

	if ocsp == nil {
		ocsp = []byte{}
	}

	// 限制长度
	if len(errString) > 300 {
		errString = errString[:300]
	}

	var query = this.Query(tx).
		Pk(certId).
		Set("ocsp", ocsp).
		Set("ocspError", errString).
		Set("ocspIsUpdated", true).
		Set("ocspUpdatedAt", time.Now().Unix()).
		Set("ocspUpdatedVersion", version).
		Set("ocspExpiresAt", expiresAt)

	if hasErr {
		query.Set("ocspTries", dbs.SQL("ocspTries+1"))
	} else {
		query.Set("ocspTries", 0)
	}

	err = query.UpdateQuickly()
	if err != nil {
		return err
	}

	// 注意：这里不通知更新，避免频繁的更新导致服务不稳定
	return nil
}

// CountAllSSLCertsWithOCSPError 计算有OCSP错误的证书数量
func (this *SSLCertDAO) CountAllSSLCertsWithOCSPError(tx *dbs.Tx, keyword string) (int64, error) {
	var query = this.Query(tx)

	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword OR dnsNames LIKE :keyword OR commonNames LIKE :keyword OR ocspError LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}

	return query.
		State(SSLCertStateEnabled).
		Attr("ocspIsUpdated", true).
		Where("LENGTH(ocspError) > 0").
		Count()
}

// ListSSLCertsWithOCSPError 列出有OCSP错误的证书
func (this *SSLCertDAO) ListSSLCertsWithOCSPError(tx *dbs.Tx, keyword string, offset int64, size int64) (result []*SSLCert, err error) {
	var query = this.Query(tx)

	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR description LIKE :keyword OR dnsNames LIKE :keyword OR commonNames LIKE :keyword OR ocspError LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}

	_, err = query.
		State(SSLCertStateEnabled).
		Attr("ocspIsUpdated", true).
		Where("LENGTH(ocspError) > 0").
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// IgnoreSSLCertsWithOCSPError 忽略一组OCSP证书错误
func (this *SSLCertDAO) IgnoreSSLCertsWithOCSPError(tx *dbs.Tx, certIds []int64) error {
	for _, certId := range certIds {
		err := this.Query(tx).
			Pk(certId).
			Set("ocspError", "").
			UpdateQuickly()
		if err != nil {
			return err
		}
	}
	return nil
}

// ResetSSLCertsWithOCSPError 重置一组证书OCSP错误状态
func (this *SSLCertDAO) ResetSSLCertsWithOCSPError(tx *dbs.Tx, certIds []int64) error {
	for _, certId := range certIds {
		err := this.Query(tx).
			Pk(certId).
			Set("ocspIsUpdated", 0).
			Set("ocspUpdatedAt", 0).
			Set("ocspError", "").
			Set("ocspTries", 0).
			UpdateQuickly()
		if err != nil {
			return err
		}
	}
	return nil
}

// ResetAllSSLCertsWithOCSPError 重置所有证书OCSP错误状态
func (this *SSLCertDAO) ResetAllSSLCertsWithOCSPError(tx *dbs.Tx) error {
	return this.Query(tx).
		State(SSLCertStateEnabled).
		Attr("ocspIsUpdated", 1).
		Where("LENGTH(ocspError)>0").
		Set("ocspIsUpdated", 0).
		Set("ocspUpdatedAt", 0).
		Set("ocspError", "").
		Set("ocspTries", 0).
		UpdateQuickly()
}

// NotifyUpdate 通知更新
func (this *SSLCertDAO) NotifyUpdate(tx *dbs.Tx, certId int64) error {
	policyIds, err := SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, certId)
	if err != nil {
		return err
	}
	if len(policyIds) == 0 {
		return nil
	}

	// 通知服务更新
	serverIds, err := SharedServerDAO.FindAllEnabledServerIdsWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return err
	}
	if len(serverIds) == 0 {
		return nil
	}
	for _, serverId := range serverIds {
		err := SharedServerDAO.NotifyUpdate(tx, serverId)
		if err != nil {
			return err
		}
	}

	// TODO 通知用户节点、API节点、管理系统（将来实现选择）更新

	return nil
}
