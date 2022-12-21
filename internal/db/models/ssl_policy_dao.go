package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

const (
	SSLPolicyStateEnabled  = 1 // 已启用
	SSLPolicyStateDisabled = 0 // 已禁用
)

type SSLPolicyDAO dbs.DAO

func NewSSLPolicyDAO() *SSLPolicyDAO {
	return dbs.NewDAO(&SSLPolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeSSLPolicies",
			Model:  new(SSLPolicy),
			PkName: "id",
		},
	}).(*SSLPolicyDAO)
}

var SharedSSLPolicyDAO *SSLPolicyDAO

func init() {
	dbs.OnReady(func() {
		SharedSSLPolicyDAO = NewSSLPolicyDAO()
	})
}

// Init 初始化
func (this *SSLPolicyDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableSSLPolicy 启用条目
func (this *SSLPolicyDAO) EnableSSLPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", SSLPolicyStateEnabled).
		Update()
	return err
}

// DisableSSLPolicy 禁用条目
func (this *SSLPolicyDAO) DisableSSLPolicy(tx *dbs.Tx, policyId int64) error {
	_, err := this.Query(tx).
		Pk(policyId).
		Set("state", SSLPolicyStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// FindEnabledSSLPolicy 查找启用中的条目
func (this *SSLPolicyDAO) FindEnabledSSLPolicy(tx *dbs.Tx, id int64) (*SSLPolicy, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", SSLPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*SSLPolicy), err
}

// ComposePolicyConfig 组合配置
func (this *SSLPolicyDAO) ComposePolicyConfig(tx *dbs.Tx, policyId int64, cacheMap *utils.CacheMap) (*sslconfigs.SSLPolicy, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}

	var cacheKey = this.Table + ":config:" + types.String(policyId)
	var cacheConfig, _ = cacheMap.Get(cacheKey)
	if cacheConfig != nil {
		return cacheConfig.(*sslconfigs.SSLPolicy), nil
	}

	policy, err := this.FindEnabledSSLPolicy(tx, policyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	config := &sslconfigs.SSLPolicy{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn
	config.ClientAuthType = int(policy.ClientAuthType)
	config.HTTP2Enabled = policy.Http2Enabled == 1
	config.MinVersion = policy.MinVersion

	// certs
	if IsNotNull(policy.Certs) {
		refs := []*sslconfigs.SSLCertRef{}
		err = json.Unmarshal(policy.Certs, &refs)
		if err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			for _, ref := range refs {
				certConfig, err := SharedSSLCertDAO.ComposeCertConfig(tx, ref.CertId, cacheMap)
				if err != nil {
					return nil, err
				}
				if certConfig == nil {
					continue
				}
				config.CertRefs = append(config.CertRefs, ref)
				config.Certs = append(config.Certs, certConfig)
			}
		}
	}

	// client CA certs
	if IsNotNull(policy.ClientCACerts) {
		refs := []*sslconfigs.SSLCertRef{}
		err = json.Unmarshal(policy.ClientCACerts, &refs)
		if err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			for _, ref := range refs {
				certConfig, err := SharedSSLCertDAO.ComposeCertConfig(tx, ref.CertId, cacheMap)
				if err != nil {
					return nil, err
				}
				if certConfig == nil {
					continue
				}
				config.ClientCARefs = append(config.ClientCARefs, ref)
				config.ClientCACerts = append(config.ClientCACerts, certConfig)
			}
		}
	}

	// cipher suites
	config.CipherSuitesIsOn = policy.CipherSuitesIsOn == 1
	if IsNotNull(policy.CipherSuites) {
		cipherSuites := []string{}
		err = json.Unmarshal(policy.CipherSuites, &cipherSuites)
		if err != nil {
			return nil, err
		}
		config.CipherSuites = cipherSuites
	}

	// hsts
	if IsNotNull(policy.Hsts) {
		hstsConfig := &sslconfigs.HSTSConfig{}
		err = json.Unmarshal(policy.Hsts, hstsConfig)
		if err != nil {
			return nil, err
		}
		config.HSTS = hstsConfig
	}

	// ocsp
	config.OCSPIsOn = policy.OcspIsOn == 1

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// FindAllEnabledPolicyIdsWithCertId 查询使用单个证书的所有策略ID
func (this *SSLPolicyDAO) FindAllEnabledPolicyIdsWithCertId(tx *dbs.Tx, certId int64) (policyIds []int64, err error) {
	if certId <= 0 {
		return
	}

	ones, err := this.Query(tx).
		State(SSLPolicyStateEnabled).
		ResultPk().
		Where("JSON_CONTAINS(certs, :certJSON)").
		Param("certJSON", maps.Map{"certId": certId}.AsJSON()).
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		policyIds = append(policyIds, int64(one.(*SSLPolicy).Id))
	}
	return policyIds, nil
}

// CreatePolicy 创建Policy
func (this *SSLPolicyDAO) CreatePolicy(tx *dbs.Tx, adminId int64, userId int64, http2Enabled bool, minVersion string, certsJSON []byte, hstsJSON []byte, ocspIsOn bool, clientAuthType int32, clientCACertsJSON []byte, cipherSuitesIsOn bool, cipherSuites []string) (int64, error) {
	var op = NewSSLPolicyOperator()
	op.State = SSLPolicyStateEnabled
	op.IsOn = true
	op.AdminId = adminId
	op.UserId = userId

	op.Http2Enabled = http2Enabled
	op.MinVersion = minVersion

	if len(certsJSON) > 0 {
		op.Certs = certsJSON
	}
	if len(hstsJSON) > 0 {
		op.Hsts = hstsJSON
	}

	op.OcspIsOn = ocspIsOn

	op.ClientAuthType = clientAuthType
	if len(clientCACertsJSON) > 0 {
		op.ClientCACerts = clientCACertsJSON
	}

	op.CipherSuitesIsOn = cipherSuitesIsOn
	if len(cipherSuites) > 0 {
		cipherSuitesJSON, err := json.Marshal(cipherSuites)
		if err != nil {
			return 0, err
		}
		op.CipherSuites = cipherSuitesJSON
	}
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdatePolicy 修改Policy
func (this *SSLPolicyDAO) UpdatePolicy(tx *dbs.Tx, policyId int64, http2Enabled bool, minVersion string, certsJSON []byte, hstsJSON []byte, ocspIsOn bool, clientAuthType int32, clientCACertsJSON []byte, cipherSuitesIsOn bool, cipherSuites []string) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	var op = NewSSLPolicyOperator()
	op.Id = policyId
	op.Http2Enabled = http2Enabled
	op.MinVersion = minVersion

	if len(certsJSON) > 0 {
		op.Certs = certsJSON
	}
	if len(hstsJSON) > 0 {
		op.Hsts = hstsJSON
	}

	op.OcspIsOn = ocspIsOn

	op.ClientAuthType = clientAuthType
	if len(clientCACertsJSON) > 0 {
		op.ClientCACerts = clientCACertsJSON
	}

	op.CipherSuitesIsOn = cipherSuitesIsOn
	if len(cipherSuites) > 0 {
		cipherSuitesJSON, err := json.Marshal(cipherSuites)
		if err != nil {
			return err
		}
		op.CipherSuites = cipherSuitesJSON
	} else {
		op.CipherSuites = "[]"
	}
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// CheckUserPolicy 检查是否为用户所属策略
func (this *SSLPolicyDAO) CheckUserPolicy(tx *dbs.Tx, userId int64, policyId int64) error {
	if policyId <= 0 || userId <= 0 {
		return ErrNotFound
	}
	ok, err := this.Query(tx).
		State(SSLPolicyStateEnabled).
		Pk(policyId).
		Attr("userId", userId).
		Exist()
	if err != nil {
		return err
	}
	if !ok {
		// 是否为当前用户的某个服务所用
		exists, err := SharedServerDAO.ExistEnabledUserServerWithSSLPolicyId(tx, userId, policyId)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}
	}
	return nil
}

// UpdatePolicyUser 修改策略所属用户
func (this *SSLPolicyDAO) UpdatePolicyUser(tx *dbs.Tx, policyId int64, userId int64) error {
	if policyId <= 0 || userId <= 0 {
		return nil
	}

	return this.Query(tx).
		Pk(policyId).
		Set("userId", userId).
		UpdateQuickly()
}

// NotifyUpdate 通知更新
func (this *SSLPolicyDAO) NotifyUpdate(tx *dbs.Tx, policyId int64) error {
	serverIds, err := SharedServerDAO.FindAllEnabledServerIdsWithSSLPolicyIds(tx, []int64{policyId})
	if err != nil {
		return err
	}
	for _, serverId := range serverIds {
		err := SharedServerDAO.NotifyUpdate(tx, serverId)
		if err != nil {
			return err
		}
	}
	return nil
}
