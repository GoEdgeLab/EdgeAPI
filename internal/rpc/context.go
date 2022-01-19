// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package rpc

import (
	"context"
	"sync"
	"time"
)

type Context struct {
	context.Context

	tagMap  map[string]time.Time
	costMap map[string]float64 // tag => costMs
	locker  sync.Mutex
}

func NewContext(ctx context.Context) *Context {
	return &Context{
		Context: ctx,
		tagMap:  map[string]time.Time{},
		costMap: map[string]float64{},
	}
}

func (this *Context) Begin(tag string) {
	this.locker.Lock()
	this.tagMap[tag] = time.Now()
	this.locker.Unlock()
}

func (this *Context) End(tag string) {
	this.locker.Lock()
	begin, ok := this.tagMap[tag]
	if ok {
		this.costMap[tag] = time.Since(begin).Seconds() * 1000
	}
	this.locker.Unlock()
}

func (this *Context) TagMap() map[string]float64 {
	this.locker.Lock()
	defer this.locker.Unlock()
	return this.costMap
}
