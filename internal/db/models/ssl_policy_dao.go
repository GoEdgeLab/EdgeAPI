package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strconv"
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

// 初始化
func (this *SSLPolicyDAO) Init() {
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
func (this *SSLPolicyDAO) EnableSSLPolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", SSLPolicyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *SSLPolicyDAO) DisableSSLPolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", SSLPolicyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *SSLPolicyDAO) FindEnabledSSLPolicy(id int64) (*SSLPolicy, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", SSLPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*SSLPolicy), err
}

// 组合配置
func (this *SSLPolicyDAO) ComposePolicyConfig(policyId int64) (*sslconfigs.SSLPolicy, error) {
	policy, err := this.FindEnabledSSLPolicy(policyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	config := &sslconfigs.SSLPolicy{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn == 1
	config.ClientAuthType = int(policy.ClientAuthType)
	config.HTTP2Enabled = policy.Http2Enabled == 1
	config.MinVersion = policy.MinVersion

	// certs
	if IsNotNull(policy.Certs) {
		refs := []*sslconfigs.SSLCertRef{}
		err = json.Unmarshal([]byte(policy.Certs), &refs)
		if err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			for _, ref := range refs {
				certConfig, err := SharedSSLCertDAO.ComposeCertConfig(ref.CertId)
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
		err = json.Unmarshal([]byte(policy.ClientCACerts), &refs)
		if err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			for _, ref := range refs {
				certConfig, err := SharedSSLCertDAO.ComposeCertConfig(ref.CertId)
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
		err = json.Unmarshal([]byte(policy.CipherSuites), &cipherSuites)
		if err != nil {
			return nil, err
		}
		config.CipherSuites = cipherSuites
	}

	// hsts
	if IsNotNull(policy.Hsts) {
		hstsConfig := &sslconfigs.HSTSConfig{}
		err = json.Unmarshal([]byte(policy.Hsts), hstsConfig)
		if err != nil {
			return nil, err
		}
		config.HSTS = hstsConfig
	}

	return config, nil
}

// 查询使用单个证书的所有策略ID
func (this *SSLPolicyDAO) FindAllEnabledPolicyIdsWithCertId(certId int64) (policyIds []int64, err error) {
	if certId <= 0 {
		return
	}

	ones, err := this.Query().
		State(SSLPolicyStateEnabled).
		ResultPk().
		Where(`JSON_CONTAINS(certs, '{"certId": ` + strconv.FormatInt(certId, 10) + ` }')`).
		Reuse(false). // 由于我们在JSON_CONTAINS()直接使用了变量，所以不能重用
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		policyIds = append(policyIds, int64(one.(*SSLPolicy).Id))
	}
	return policyIds, nil
}

// 创建Policy
func (this *SSLPolicyDAO) CreatePolicy(http2Enabled bool, minVersion string, certsJSON []byte, hstsJSON []byte, clientAuthType int32, clientCACertsJSON []byte, cipherSuitesIsOn bool, cipherSuites []string) (int64, error) {
	op := NewSSLPolicyOperator()
	op.State = SSLPolicyStateEnabled
	op.IsOn = true
	op.Http2Enabled = http2Enabled
	op.MinVersion = minVersion

	if len(certsJSON) > 0 {
		op.Certs = certsJSON
	}
	if len(hstsJSON) > 0 {
		op.Hsts = hstsJSON
	}

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
	err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改Policy
// 创建Policy
func (this *SSLPolicyDAO) UpdatePolicy(policyId int64, http2Enabled bool, minVersion string, certsJSON []byte, hstsJSON []byte, clientAuthType int32, clientCACertsJSON []byte, cipherSuitesIsOn bool, cipherSuites []string) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	op := NewSSLPolicyOperator()
	op.Id = policyId
	op.Http2Enabled = http2Enabled
	op.MinVersion = minVersion

	if len(certsJSON) > 0 {
		op.Certs = certsJSON
	}
	if len(hstsJSON) > 0 {
		op.Hsts = hstsJSON
	}

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
	err := this.Save(op)
	return err
}
