package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPHeaderPolicyStateEnabled  = 1 // 已启用
	HTTPHeaderPolicyStateDisabled = 0 // 已禁用
)

type HTTPHeaderPolicyDAO dbs.DAO

func NewHTTPHeaderPolicyDAO() *HTTPHeaderPolicyDAO {
	return dbs.NewDAO(&HTTPHeaderPolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPHeaderPolicies",
			Model:  new(HTTPHeaderPolicy),
			PkName: "id",
		},
	}).(*HTTPHeaderPolicyDAO)
}

var SharedHTTPHeaderPolicyDAO *HTTPHeaderPolicyDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPHeaderPolicyDAO = NewHTTPHeaderPolicyDAO()
	})
}

// Init 初始化
func (this *HTTPHeaderPolicyDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPHeaderPolicy 启用条目
func (this *HTTPHeaderPolicyDAO) EnableHTTPHeaderPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPHeaderPolicyStateEnabled).
		Update()
	return err
}

// DisableHTTPHeaderPolicy 禁用条目
func (this *HTTPHeaderPolicyDAO) DisableHTTPHeaderPolicy(tx *dbs.Tx, policyId int64) error {
	_, err := this.Query(tx).
		Pk(policyId).
		Set("state", HTTPHeaderPolicyStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// FindEnabledHTTPHeaderPolicy 查找启用中的条目
func (this *HTTPHeaderPolicyDAO) FindEnabledHTTPHeaderPolicy(tx *dbs.Tx, id int64) (*HTTPHeaderPolicy, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPHeaderPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPHeaderPolicy), err
}

// CreateHeaderPolicy 创建策略
func (this *HTTPHeaderPolicyDAO) CreateHeaderPolicy(tx *dbs.Tx) (int64, error) {
	var op = NewHTTPHeaderPolicyOperator()
	op.IsOn = true
	op.State = HTTPHeaderPolicyStateEnabled
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateAddingHeaders 修改AddHeaders
func (this *HTTPHeaderPolicyDAO) UpdateAddingHeaders(tx *dbs.Tx, policyId int64, headersJSON []byte) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	var op = NewHTTPHeaderPolicyOperator()
	op.Id = policyId
	op.AddHeaders = headersJSON
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// UpdateSettingHeaders 修改SetHeaders
func (this *HTTPHeaderPolicyDAO) UpdateSettingHeaders(tx *dbs.Tx, policyId int64, headersJSON []byte) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	var op = NewHTTPHeaderPolicyOperator()
	op.Id = policyId
	op.SetHeaders = headersJSON
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// UpdateReplacingHeaders 修改ReplaceHeaders
func (this *HTTPHeaderPolicyDAO) UpdateReplacingHeaders(tx *dbs.Tx, policyId int64, headersJSON []byte) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	var op = NewHTTPHeaderPolicyOperator()
	op.Id = policyId
	op.ReplaceHeaders = headersJSON
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// UpdateAddingTrailers 修改AddTrailers
func (this *HTTPHeaderPolicyDAO) UpdateAddingTrailers(tx *dbs.Tx, policyId int64, headersJSON []byte) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	var op = NewHTTPHeaderPolicyOperator()
	op.Id = policyId
	op.AddTrailers = headersJSON
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// UpdateDeletingHeaders 修改DeleteHeaders
func (this *HTTPHeaderPolicyDAO) UpdateDeletingHeaders(tx *dbs.Tx, policyId int64, headerNames []string) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	if headerNames == nil {
		headerNames = []string{}
	}
	namesJSON, err := json.Marshal(headerNames)
	if err != nil {
		return err
	}

	var op = NewHTTPHeaderPolicyOperator()
	op.Id = policyId
	op.DeleteHeaders = namesJSON
	err = this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// UpdateNonStandardHeaders 修改非标Headers
func (this *HTTPHeaderPolicyDAO) UpdateNonStandardHeaders(tx *dbs.Tx, policyId int64, headerNames []string) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}

	if headerNames == nil {
		headerNames = []string{}
	}
	namesJSON, err := json.Marshal(headerNames)
	if err != nil {
		return err
	}

	var op = NewHTTPHeaderPolicyOperator()
	op.Id = policyId
	op.NonStandardHeaders = namesJSON
	err = this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, policyId)
}

// ComposeHeaderPolicyConfig 组合配置
func (this *HTTPHeaderPolicyDAO) ComposeHeaderPolicyConfig(tx *dbs.Tx, headerPolicyId int64) (*shared.HTTPHeaderPolicy, error) {
	policy, err := this.FindEnabledHTTPHeaderPolicy(tx, headerPolicyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}

	var config = &shared.HTTPHeaderPolicy{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn

	// SetHeaders
	if IsNotNull(policy.SetHeaders) {
		var refs = []*shared.HTTPHeaderRef{}
		err = json.Unmarshal(policy.SetHeaders, &refs)
		if err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			var resultRefs = []*shared.HTTPHeaderRef{}
			for _, ref := range refs {
				headerConfig, err := SharedHTTPHeaderDAO.ComposeHeaderConfig(tx, ref.HeaderId)
				if err != nil {
					return nil, err
				}
				if headerConfig == nil {
					continue
				}
				resultRefs = append(resultRefs, ref)
				config.SetHeaders = append(config.SetHeaders, headerConfig)
			}
			config.SetHeaderRefs = resultRefs
		}
	}

	// Delete Headers
	if IsNotNull(policy.DeleteHeaders) {
		var headers = []string{}
		err = json.Unmarshal(policy.DeleteHeaders, &headers)
		if err != nil {
			return nil, err
		}
		config.DeleteHeaders = headers
	}

	// Non-Standard Headers
	if IsNotNull(policy.NonStandardHeaders) {
		var headers = []string{}
		err = json.Unmarshal(policy.NonStandardHeaders, &headers)
		if err != nil {
			return nil, err
		}
		config.NonStandardHeaders = headers
	}

	// CORS
	if IsNotNull(policy.Cors) {
		var corsConfig = shared.NewHTTPCORSHeaderConfig()
		err = json.Unmarshal(policy.Cors, corsConfig)
		if err != nil {
			return nil, err
		}
		config.CORS = corsConfig
	}

	// Expires
	// TODO

	return config, nil
}

// FindHeaderPolicyIdWithHeaderId 查找Header所在Policy
func (this *HTTPHeaderPolicyDAO) FindHeaderPolicyIdWithHeaderId(tx *dbs.Tx, headerId int64) (int64, error) {
	return this.Query(tx).
		Where("(JSON_CONTAINS(addHeaders, :jsonQuery) OR JSON_CONTAINS(addTrailers, :jsonQuery) OR JSON_CONTAINS(setHeaders, :jsonQuery) OR JSON_CONTAINS(replaceHeaders, :jsonQuery))").
		Param("jsonQuery", maps.Map{"headerId": headerId}.AsJSON()).
		ResultPk().
		FindInt64Col(0)
}

// UpdateHeaderPolicyCORS 修改CORS
func (this *HTTPHeaderPolicyDAO) UpdateHeaderPolicyCORS(tx *dbs.Tx, headerPolicyId int64, corsConfig *shared.HTTPCORSHeaderConfig) error {
	if headerPolicyId <= 0 {
		return errors.New("invalid headerId")
	}

	corsJSON, err := json.Marshal(corsConfig)
	if err != nil {
		return err
	}

	err = this.Query(tx).
		Pk(headerPolicyId).
		Set("cors", corsJSON).
		UpdateQuickly()
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, headerPolicyId)
}

// CheckUserHeaderPolicy 检查用户权限
func (this *HTTPHeaderPolicyDAO) CheckUserHeaderPolicy(tx *dbs.Tx, userId int64, policyId int64) error {
	if userId <= 0 || policyId <= 0 {
		return ErrNotFound
	}

	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithHeaderPolicyId(tx, policyId)
	if err != nil {
		return err
	}

	if webId <= 0 {
		return ErrNotFound
	}

	return SharedHTTPWebDAO.CheckUserWeb(tx, userId, webId)
}

// NotifyUpdate 通知更新
func (this *HTTPHeaderPolicyDAO) NotifyUpdate(tx *dbs.Tx, policyId int64) error {
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithHeaderPolicyId(tx, policyId)
	if err != nil {
		return err
	}
	if webId > 0 {
		return SharedHTTPWebDAO.NotifyUpdate(tx, webId)
	}
	return nil
}
