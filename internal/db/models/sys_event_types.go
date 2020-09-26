package models

import (
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
	serverIds, err := SharedServerDAO.FindAllEnabledServerIds()
	if err != nil {
		return err
	}
	for _, serverId := range serverIds {
		isChanged, err := SharedServerDAO.RenewServerConfig(serverId)
		if err != nil {
			return err
		}
		if !isChanged {
			continue
		}

		// 检查节点是否需要更新
		isOk, clusterId, err := SharedServerDAO.FindServerNodeFilters(serverId)
		if err != nil {
			return err
		}
		if !isOk {
			continue
		}
		err = SharedNodeDAO.UpdateAllNodesLatestVersionMatch(clusterId)
		if err != nil {
			return err
		}
	}

	return nil
}
