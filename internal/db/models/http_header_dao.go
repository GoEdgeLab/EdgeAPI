package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPHeaderStateEnabled  = 1 // 已启用
	HTTPHeaderStateDisabled = 0 // 已禁用
)

type HTTPHeaderDAO dbs.DAO

func NewHTTPHeaderDAO() *HTTPHeaderDAO {
	return dbs.NewDAO(&HTTPHeaderDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPHeaders",
			Model:  new(HTTPHeader),
			PkName: "id",
		},
	}).(*HTTPHeaderDAO)
}

var SharedHTTPHeaderDAO *HTTPHeaderDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPHeaderDAO = NewHTTPHeaderDAO()
	})
}

// Init 初始化
func (this *HTTPHeaderDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPHeader 启用条目
func (this *HTTPHeaderDAO) EnableHTTPHeader(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPHeaderStateEnabled).
		Update()
	return err
}

// DisableHTTPHeader 禁用条目
func (this *HTTPHeaderDAO) DisableHTTPHeader(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPHeaderStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPHeader 查找启用中的条目
func (this *HTTPHeaderDAO) FindEnabledHTTPHeader(tx *dbs.Tx, id int64) (*HTTPHeader, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPHeaderStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPHeader), err
}

// FindHTTPHeaderName 根据主键查找名称
func (this *HTTPHeaderDAO) FindHTTPHeaderName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateHeader 创建Header
func (this *HTTPHeaderDAO) CreateHeader(tx *dbs.Tx, userId int64, name string, value string, status []int, disableRedirect bool, shouldAppend bool, shouldReplace bool, replaceValues []*shared.HTTPHeaderReplaceValue, methods []string, domains []string) (int64, error) {
	var op = NewHTTPHeaderOperator()
	op.UserId = userId
	op.State = HTTPHeaderStateEnabled
	op.IsOn = true
	op.Name = name
	op.Value = value

	// status
	var statusConfig *shared.HTTPStatusConfig
	if len(status) == 0 {
		statusConfig = &shared.HTTPStatusConfig{
			Always: true,
		}
	} else {
		statusConfig = &shared.HTTPStatusConfig{
			Always: false,
			Codes:  status,
		}
	}

	statusJSON, err := json.Marshal(statusConfig)
	if err != nil {
		return 0, err
	}
	op.Status = statusJSON

	op.DisableRedirect = disableRedirect
	op.ShouldAppend = shouldAppend
	op.ShouldReplace = shouldReplace

	if len(replaceValues) == 0 {
		op.ReplaceValues = "[]"
	} else {
		replaceValuesJSON, err := json.Marshal(replaceValues)
		if err != nil {
			return 0, err
		}
		op.ReplaceValues = replaceValuesJSON
	}

	// methods
	if len(methods) == 0 {
		op.Methods = "[]"
	} else {
		methodsJSON, err := json.Marshal(methods)
		if err != nil {
			return 0, err
		}
		op.Methods = methodsJSON
	}

	// domains
	if len(domains) == 0 {
		op.Domains = "[]"
	} else {
		domainsJSON, err := json.Marshal(domains)
		if err != nil {
			return 0, err
		}
		op.Domains = domainsJSON
	}

	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateHeader 修改Header
func (this *HTTPHeaderDAO) UpdateHeader(tx *dbs.Tx, headerId int64, name string, value string, status []int, disableRedirect bool, shouldAppend bool, shouldReplace bool, replaceValues []*shared.HTTPHeaderReplaceValue, methods []string, domains []string) error {
	if headerId <= 0 {
		return errors.New("invalid headerId")
	}

	var op = NewHTTPHeaderOperator()
	op.Id = headerId
	op.Name = name
	op.Value = value

	// status
	var statusConfig *shared.HTTPStatusConfig
	if len(status) == 0 {
		statusConfig = &shared.HTTPStatusConfig{
			Always: true,
		}
	} else {
		statusConfig = &shared.HTTPStatusConfig{
			Always: false,
			Codes:  status,
		}
	}

	statusJSON, err := json.Marshal(statusConfig)
	if err != nil {
		return err
	}
	op.Status = statusJSON

	op.DisableRedirect = disableRedirect
	op.ShouldAppend = shouldAppend
	op.ShouldReplace = shouldReplace

	if len(replaceValues) == 0 {
		op.ReplaceValues = "[]"
	} else {
		replaceValuesJSON, err := json.Marshal(replaceValues)
		if err != nil {
			return err
		}
		op.ReplaceValues = replaceValuesJSON
	}

	// methods
	if len(methods) == 0 {
		op.Methods = "[]"
	} else {
		methodsJSON, err := json.Marshal(methods)
		if err != nil {
			return err
		}
		op.Methods = methodsJSON
	}

	// domains
	if len(domains) == 0 {
		op.Domains = "[]"
	} else {
		domainsJSON, err := json.Marshal(domains)
		if err != nil {
			return err
		}
		op.Domains = domainsJSON
	}

	err = this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, headerId)
}

// ComposeHeaderConfig 组合Header配置
func (this *HTTPHeaderDAO) ComposeHeaderConfig(tx *dbs.Tx, headerId int64) (*shared.HTTPHeaderConfig, error) {
	header, err := this.FindEnabledHTTPHeader(tx, headerId)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, nil
	}

	config := &shared.HTTPHeaderConfig{}
	config.Id = int64(header.Id)
	config.IsOn = header.IsOn
	config.Name = header.Name
	config.Value = header.Value
	config.DisableRedirect = header.DisableRedirect == 1
	config.ShouldAppend = header.ShouldAppend

	// replace
	config.ShouldReplace = header.ShouldReplace
	if IsNotNull(header.ReplaceValues) {
		var values = []*shared.HTTPHeaderReplaceValue{}
		err = json.Unmarshal(header.ReplaceValues, &values)
		if err != nil {
			return nil, err
		}
		config.ReplaceValues = values
	}

	// status
	if IsNotNull(header.Status) {
		status := &shared.HTTPStatusConfig{}
		err = json.Unmarshal(header.Status, status)
		if err != nil {
			return nil, err
		}
		config.Status = status
	}

	// methods
	if IsNotNull(header.Methods) {
		var methods = []string{}
		err = json.Unmarshal(header.Methods, &methods)
		if err != nil {
			return nil, err
		}
		config.Methods = methods
	}

	// domains
	if IsNotNull(header.Domains) {
		var domains = []string{}
		err = json.Unmarshal(header.Domains, &domains)
		if err != nil {
			return nil, err
		}
		config.Domains = domains
	}

	return config, nil
}

// NotifyUpdate 通知更新
func (this *HTTPHeaderDAO) NotifyUpdate(tx *dbs.Tx, headerId int64) error {
	policyId, err := SharedHTTPHeaderPolicyDAO.FindHeaderPolicyIdWithHeaderId(tx, headerId)
	if err != nil {
		return err
	}
	if policyId > 0 {
		return SharedHTTPHeaderPolicyDAO.NotifyUpdate(tx, policyId)
	}
	return nil
}
