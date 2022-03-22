package models

import (
	"encoding/json"
	"errors"
	"reflect"
)

// DecodeEvent 解码事件
func (this *SysEvent) DecodeEvent() (EventInterface, error) {
	// 解析数据类型
	t, isOk := eventTypeMapping[this.Type]
	if !isOk {
		return nil, errors.New("can not found event type '" + this.Type + "'")
	}
	ptr := reflect.New(t).Interface().(EventInterface)

	// 解析参数
	if IsNotNull(this.Params) {
		err := json.Unmarshal(this.Params, ptr)
		if err != nil {
			return nil, err
		}
	}

	return ptr, nil
}
