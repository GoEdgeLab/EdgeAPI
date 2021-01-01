package models

import (
	"github.com/iwind/TeaGo/dbs"
	"reflect"
)

var eventTypeMapping = map[string]reflect.Type{} // eventType => reflect type

func init() {
	for _, event := range []EventInterface{
		NewServerChangeEvent(),
	} {
		eventTypeMapping[event.Type()] = reflect.ValueOf(event).Elem().Type()
	}
}

// 接口
type EventInterface interface {
	Type() string
	Run() error
}

// 服务变化
type ServerChangeEvent struct {
}

func NewServerChangeEvent() *ServerChangeEvent {
	return &ServerChangeEvent{}
}

func (this *ServerChangeEvent) Type() string {
	return "serverChange"
}

func (this *ServerChangeEvent) Run() error {
	var tx *dbs.Tx

	serverIds, err := SharedServerDAO.FindAllEnabledServerIds(tx)
	if err != nil {
		return err
	}
	for _, serverId := range serverIds {
		isChanged, err := SharedServerDAO.RenewServerConfig(tx, serverId, true)
		if err != nil {
			return err
		}
		if !isChanged {
			continue
		}

		// 检查节点是否需要更新
		isOk, clusterId, err := SharedServerDAO.FindServerNodeFilters(tx, serverId)
		if err != nil {
			return err
		}
		if !isOk {
			continue
		}
		err = SharedNodeDAO.IncreaseAllNodesLatestVersionMatch(tx, clusterId)
		if err != nil {
			return err
		}
	}

	return nil
}
