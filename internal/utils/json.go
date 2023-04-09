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
