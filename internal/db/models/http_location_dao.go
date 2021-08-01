package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPLocationStateEnabled  = 1 // 已启用
	HTTPLocationStateDisabled = 0 // 已禁用
)

type HTTPLocationDAO dbs.DAO

func NewHTTPLocationDAO() *HTTPLocationDAO {
	return dbs.NewDAO(&HTTPLocationDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPLocations",
			Model:  new(HTTPLocation),
			PkName: "id",
		},
	}).(*HTTPLocationDAO)
}

var SharedHTTPLocationDAO *HTTPLocationDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPLocationDAO = NewHTTPLocationDAO()
	})
}

// Init 初始化
func (this *HTTPLocationDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPLocation 启用条目
func (this *HTTPLocationDAO) EnableHTTPLocation(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPLocationStateEnabled).
		Update()
	return err
}

// DisableHTTPLocation 禁用条目
func (this *HTTPLocationDAO) DisableHTTPLocation(tx *dbs.Tx, locationId int64) error {
	_, err := this.Query(tx).
		Pk(locationId).
		Set("state", HTTPLocationStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, locationId)
}

// FindEnabledHTTPLocation 查找启用中的条目
func (this *HTTPLocationDAO) FindEnabledHTTPLocation(tx *dbs.Tx, id int64) (*HTTPLocation, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPLocationStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPLocation), err
}

// FindHTTPLocationName 根据主键查找名称
func (this *HTTPLocationDAO) FindHTTPLocationName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateLocation 创建路由规则
func (this *HTTPLocationDAO) CreateLocation(tx *dbs.Tx, parentId int64, name string, pattern string, description string, isBreak bool, condsJSON []byte) (int64, error) {
	op := NewHTTPLocationOperator()
	op.IsOn = true
	op.State = HTTPLocationStateEnabled
	op.ParentId = parentId
	op.Name = name
	op.Pattern = pattern
	op.Description = description
	op.IsBreak = isBreak

	if len(condsJSON) > 0 {
		op.Conds = condsJSON
	}

	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateLocation 修改路由规则
func (this *HTTPLocationDAO) UpdateLocation(tx *dbs.Tx, locationId int64, name string, pattern string, description string, isOn bool, isBreak bool, condsJSON []byte) error {
	if locationId <= 0 {
		return errors.New("invalid locationId")
	}
	op := NewHTTPLocationOperator()
	op.Id = locationId
	op.Name = name
	op.Pattern = pattern
	op.Description = description
	op.IsOn = isOn
	op.IsBreak = isBreak

	if len(condsJSON) > 0 {
		op.Conds = condsJSON
	}

	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, locationId)
}

// ComposeLocationConfig 组合配置
func (this *HTTPLocationDAO) ComposeLocationConfig(tx *dbs.Tx, locationId int64) (*serverconfigs.HTTPLocationConfig, error) {
	location, err := this.FindEnabledHTTPLocation(tx, locationId)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, nil
	}

	config := &serverconfigs.HTTPLocationConfig{}
	config.Id = int64(location.Id)
	config.IsOn = location.IsOn == 1
	config.Description = location.Description
	config.Name = location.Name
	config.Pattern = location.Pattern
	config.URLPrefix = location.UrlPrefix
	config.IsBreak = location.IsBreak == 1

	// web
	if location.WebId > 0 {
		webConfig, err := SharedHTTPWebDAO.ComposeWebConfig(tx, int64(location.WebId))
		if err != nil {
			return nil, err
		}
		config.Web = webConfig
	}

	// reverse proxy
	if IsNotNull(location.ReverseProxy) {
		ref := &serverconfigs.ReverseProxyRef{}
		err = json.Unmarshal([]byte(location.ReverseProxy), ref)
		if err != nil {
			return nil, err
		}
		config.ReverseProxyRef = ref
		if ref.ReverseProxyId > 0 {
			reverseProxyConfig, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, ref.ReverseProxyId)
			if err != nil {
				return nil, err
			}
			config.ReverseProxy = reverseProxyConfig
		}
	}

	// conds
	if len(location.Conds) > 0 {
		conds := &shared.HTTPRequestCondsConfig{}
		err = json.Unmarshal([]byte(location.Conds), conds)
		if err != nil {
			return nil, err
		}
		config.Conds = conds
	}

	return config, nil
}

// FindLocationReverseProxy 查找反向代理设置
func (this *HTTPLocationDAO) FindLocationReverseProxy(tx *dbs.Tx, locationId int64) (*serverconfigs.ReverseProxyRef, error) {
	refString, err := this.Query(tx).
		Pk(locationId).
		Result("reverseProxy").
		FindStringCol("")
	if err != nil {
		return nil, err
	}
	if IsNotNull(refString) {
		ref := &serverconfigs.ReverseProxyRef{}
		err = json.Unmarshal([]byte(refString), ref)
		if err != nil {
			return nil, err
		}
		return ref, nil
	}
	return nil, nil
}

// UpdateLocationReverseProxy 更改反向代理设置
func (this *HTTPLocationDAO) UpdateLocationReverseProxy(tx *dbs.Tx, locationId int64, reverseProxyJSON []byte) error {
	if locationId <= 0 {
		return errors.New("invalid locationId")
	}
	op := NewHTTPLocationOperator()
	op.Id = locationId
	op.ReverseProxy = JSONBytes(reverseProxyJSON)
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, locationId)
}

// FindLocationWebId 查找WebId
func (this *HTTPLocationDAO) FindLocationWebId(tx *dbs.Tx, locationId int64) (int64, error) {
	webId, err := this.Query(tx).
		Pk(locationId).
		Result("webId").
		FindIntCol(0)
	return int64(webId), err
}

// UpdateLocationWeb 更改Web设置
func (this *HTTPLocationDAO) UpdateLocationWeb(tx *dbs.Tx, locationId int64, webId int64) error {
	if locationId <= 0 {
		return errors.New("invalid locationId")
	}
	op := NewHTTPLocationOperator()
	op.Id = locationId
	op.WebId = webId
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, locationId)
}

// ConvertLocationRefs 转换引用为配置
func (this *HTTPLocationDAO) ConvertLocationRefs(tx *dbs.Tx, refs []*serverconfigs.HTTPLocationRef) (locations []*serverconfigs.HTTPLocationConfig, err error) {
	for _, ref := range refs {
		config, err := this.ComposeLocationConfig(tx, ref.LocationId)
		if err != nil {
			return nil, err
		}
		children, err := this.ConvertLocationRefs(tx, ref.Children)
		if err != nil {
			return nil, err
		}
		config.Children = children
		locations = append(locations, config)
	}

	return
}

// FindEnabledLocationIdWithWebId 根据WebId查找LocationId
func (this *HTTPLocationDAO) FindEnabledLocationIdWithWebId(tx *dbs.Tx, webId int64) (locationId int64, err error) {
	if webId <= 0 {
		return
	}
	return this.Query(tx).
		Attr("webId", webId).
		ResultPk().
		FindInt64Col(0)
}

// FindEnabledLocationIdWithReverseProxyId 查找包含某个反向代理的Server
func (this *HTTPLocationDAO) FindEnabledLocationIdWithReverseProxyId(tx *dbs.Tx, reverseProxyId int64) (serverId int64, err error) {
	return this.Query(tx).
		State(ServerStateEnabled).
		Where("JSON_CONTAINS(reverseProxy, :jsonQuery)").
		Param("jsonQuery", maps.Map{"reverseProxyId": reverseProxyId}.AsJSON()).
		ResultPk().
		FindInt64Col(0)
}


// NotifyUpdate 通知更新
func (this *HTTPLocationDAO) NotifyUpdate(tx *dbs.Tx, locationId int64) error {
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithLocationId(tx, locationId)
	if err != nil {
		return err
	}
	if webId > 0 {
		return SharedHTTPWebDAO.NotifyUpdate(tx, webId)
	}
	return nil
}
