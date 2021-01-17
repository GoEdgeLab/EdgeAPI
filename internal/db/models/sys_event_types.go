package models

import (
	"reflect"
)

var eventTypeMapping = map[string]reflect.Type{} // eventType => reflect type

func init() {
	for _, event := range []EventInterface{
		// Event列表
	} {
		eventTypeMapping[event.Type()] = reflect.ValueOf(event).Elem().Type()
	}
}

// 接口
type EventInterface interface {
	Type() string
	Run() error
}
