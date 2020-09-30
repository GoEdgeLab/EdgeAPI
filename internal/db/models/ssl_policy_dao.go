package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
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

var SharedSSLPolicyDAO = NewSSLPolicyDAO()

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

	// cipher suites
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
