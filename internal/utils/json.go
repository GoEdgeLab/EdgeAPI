// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils

import (
	"encoding/json"
	"errors"
	"reflect"
)

// JSONClone 使用JSON协议克隆对象
func JSONClone[T any](ptr T) (newPtr T, err error) {
	var ptrType = reflect.TypeOf(ptr)
	var kind = ptrType.Kind()
	if kind != reflect.Ptr && kind != reflect.Slice {
		err = errors.New("JSONClone: input must be a ptr or slice")
		return
	}
	var jsonData []byte
	jsonData, err = json.Marshal(ptr)
	if err != nil {
		return ptr, errors.New("JSONClone: marshal failed: " + err.Error())
	}

	var newValue any
	switch kind {
	case reflect.Ptr:
		newValue = reflect.New(ptrType.Elem()).Interface()
	case reflect.Slice:
		newValue = reflect.New(reflect.SliceOf(ptrType.Elem())).Interface()
	default:
		return ptr, errors.New("JSONClone: unknown data type")
	}
	err = json.Unmarshal(jsonData, newValue)
	if err != nil {
		err = errors.New("JSONClone: unmarshal failed: " + err.Error())
		return
	}

	if kind == reflect.Slice {
		newValue = reflect.Indirect(reflect.ValueOf(newValue)).Interface()
	}

	return newValue.(T), nil
}

// JSONDecodeConfig 解码并重新编码
// 是为了去除原有JSON中不需要的数据
func JSONDecodeConfig(data []byte, ptr any) (encodeJSON []byte, err error) {
	err = json.Unmarshal(data, ptr)
	if err != nil {
		return
	}

	encodeJSON, err = json.Marshal(ptr)
	if err != nil {
		return
	}

	// validate config
	if ptr != nil {
		config, ok := ptr.(interface {
			Init() error
		})
		if ok {
			initErr := config.Init()
			if initErr != nil {
				err = errors.New("validate config failed: " + initErr.Error())
			}
		}
	}

	return
}
