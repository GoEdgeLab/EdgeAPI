// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import (
	"github.com/iwind/TeaGo/maps"
	"sync"
)

type CacheMap struct {
	locker sync.Mutex
	m      maps.Map
}

func NewCacheMap() *CacheMap {
	return &CacheMap{m: maps.Map{}}
}

func (this *CacheMap) Get(key string) (value interface{}, ok bool) {
	this.locker.Lock()
	value, ok = this.m[key]
	this.locker.Unlock()
	return
}

func (this *CacheMap) Put(key string, value interface{}) {
	if value == nil {
		return
	}
	this.locker.Lock()
	this.m[key] = value
	this.locker.Unlock()
}

func (this *CacheMap) Len() int {
	this.locker.Lock()
	var l = len(this.m)
	this.locker.Unlock()
	return l
}
