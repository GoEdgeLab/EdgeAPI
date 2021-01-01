package models

import (
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPWebsocketStateEnabled  = 1 // 已启用
	HTTPWebsocketStateDisabled = 0 // 已禁用
)

type HTTPWebsocketDAO dbs.DAO

func NewHTTPWebsocketDAO() *HTTPWebsocketDAO {
	return dbs.NewDAO(&HTTPWebsocketDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPWebsockets",
			Model:  new(HTTPWebsocket),
			PkName: "id",
		},
	}).(*HTTPWebsocketDAO)
}

var SharedHTTPWebsocketDAO *HTTPWebsocketDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPWebsocketDAO = NewHTTPWebsocketDAO()
	})
}

// 启用条目
func (this *HTTPWebsocketDAO) EnableHTTPWebsocket(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPWebsocketStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPWebsocketDAO) DisableHTTPWebsocket(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPWebsocketStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPWebsocketDAO) FindEnabledHTTPWebsocket(tx *dbs.Tx, id int64) (*HTTPWebsocket, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPWebsocketStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPWebsocket), err
}

// 组合配置
func (this *HTTPWebsocketDAO) ComposeWebsocketConfig(tx *dbs.Tx, websocketId int64) (*serverconfigs.HTTPWebsocketConfig, error) {
	websocket, err := this.FindEnabledHTTPWebsocket(tx, websocketId)
	if err != nil {
		return nil, err
	}
	if websocket == nil {
		return nil, nil
	}
	config := &serverconfigs.HTTPWebsocketConfig{}
	config.Id = int64(websocket.Id)
	config.IsOn = websocket.IsOn == 1
	config.AllowAllOrigins = websocket.AllowAllOrigins == 1

	if IsNotNull(websocket.AllowedOrigins) {
		origins := []string{}
		err = json.Unmarshal([]byte(websocket.AllowedOrigins), &origins)
		if err != nil {
			return nil, err
		}
		config.AllowedOrigins = origins
	}

	if IsNotNull(websocket.HandshakeTimeout) {
		duration := &shared.TimeDuration{}
		err = json.Unmarshal([]byte(websocket.HandshakeTimeout), duration)
		if err != nil {
			return nil, err
		}
		config.HandshakeTimeout = duration
	}

	config.RequestSameOrigin = websocket.RequestSameOrigin == 1
	config.RequestOrigin = websocket.RequestOrigin

	return config, nil
}

// 创建Websocket配置
func (this *HTTPWebsocketDAO) CreateWebsocket(tx *dbs.Tx, handshakeTimeoutJSON []byte, allowAllOrigins bool, allowedOrigins []string, requestSameOrigin bool, requestOrigin string) (websocketId int64, err error) {
	op := NewHTTPWebsocketOperator()
	op.IsOn = true
	op.State = HTTPWebsocketStateEnabled
	if len(handshakeTimeoutJSON) > 0 {
		op.HandshakeTimeout = handshakeTimeoutJSON
	}
	op.AllowAllOrigins = allowAllOrigins
	if len(allowedOrigins) > 0 {
		originsJSON, err := json.Marshal(allowedOrigins)
		if err != nil {
			return 0, err
		}
		op.AllowedOrigins = originsJSON
	}
	op.RequestSameOrigin = requestSameOrigin
	op.RequestOrigin = requestOrigin
	err = this.Save(tx, op)
	return types.Int64(op.Id), err
}

// 修改Websocket配置
func (this *HTTPWebsocketDAO) UpdateWebsocket(tx *dbs.Tx, websocketId int64, handshakeTimeoutJSON []byte, allowAllOrigins bool, allowedOrigins []string, requestSameOrigin bool, requestOrigin string) error {
	if websocketId <= 0 {
		return errors.New("invalid websocketId")
	}
	op := NewHTTPWebsocketOperator()
	op.Id = websocketId
	if len(handshakeTimeoutJSON) > 0 {
		op.HandshakeTimeout = handshakeTimeoutJSON
	}
	op.AllowAllOrigins = allowAllOrigins
	if len(allowedOrigins) > 0 {
		originsJSON, err := json.Marshal(allowedOrigins)
		if err != nil {
			return err
		}
		op.AllowedOrigins = originsJSON
	} else {
		op.AllowedOrigins = "[]"
	}
	op.RequestSameOrigin = requestSameOrigin
	op.RequestOrigin = requestOrigin
	err := this.Save(tx, op)
	return err
}
