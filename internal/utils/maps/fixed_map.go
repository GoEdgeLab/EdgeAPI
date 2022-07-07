// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package maputils

import (
	"sync"
)

type FixedMap struct {
	m    map[string]any
	keys []string

	maxSize int

	locker sync.RWMutex
}

func NewFixedMap(maxSize int) *FixedMap {
	return &FixedMap{
		m:       map[string]any{},
		maxSize: maxSize,
	}
}

func (this *FixedMap) Set(key string, item any) {
	if this.maxSize <= 0 {
		return
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	_, ok := this.m[key]
	if ok {
		this.m[key] = item

		// TODO 将key转到keys末尾
	} else {
		// 是否已满
		if len(this.keys) >= this.maxSize {
			var firstKey = this.keys[0]
			delete(this.m, firstKey)
			this.keys = this.keys[1:]
		}

		// 新加入
		this.m[key] = item
		this.keys = append(this.keys, key)
	}
}

func (this *FixedMap) Get(key string) (value any, ok bool) {
	this.locker.RLock()
	value, ok = this.m[key]
	this.locker.RUnlock()
	return
}

func (this *FixedMap) Has(key string) bool {
	this.locker.RLock()
	_, ok := this.m[key]
	this.locker.RUnlock()
	return ok
}

func (this *FixedMap) Size() int {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return len(this.keys)
}

func (this *FixedMap) Reset() {
	this.locker.Lock()
	this.m = map[string]any{}
	this.keys = []string{}
	this.locker.Unlock()
}
