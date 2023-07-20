// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package models

import (
	"sync"
)

type SysLockerIncrementItem struct {
	size     int
	c        chan int64
	maxValue int64
}

func NewSysLockerIncrementItem(size int) *SysLockerIncrementItem {
	if size <= 0 {
		size = 10
	}

	return &SysLockerIncrementItem{
		size: size,
		c:    make(chan int64, size),
	}
}

func (this *SysLockerIncrementItem) Pop() (result int64, ok bool) {
	select {
	case v := <-this.c:
		result = v
		ok = true
		return
	default:
		return
	}
}

func (this *SysLockerIncrementItem) Push(value int64) {
	if this.maxValue < value {
		this.maxValue = value
	}

	select {
	case this.c <- value:
	default:
	}
}

func (this *SysLockerIncrementItem) Reset() {
	close(this.c)
	this.c = make(chan int64, this.size)
}

func (this *SysLockerIncrementItem) MaxValue() int64 {
	return this.maxValue
}

type SysLockerIncrement struct {
	itemMap map[string]*SysLockerIncrementItem // key => item
	size    int
	locker  sync.RWMutex
}

func NewSysLockerIncrement(size int) *SysLockerIncrement {
	if size <= 0 {
		size = 10
	}

	return &SysLockerIncrement{
		itemMap: map[string]*SysLockerIncrementItem{},
		size:    size,
	}
}

func (this *SysLockerIncrement) Pop(key string) (result int64, ok bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	item, itemOk := this.itemMap[key]
	if itemOk {
		result, ok = item.Pop()
	}
	return
}

func (this *SysLockerIncrement) Push(key string, minValue int64, maxValue int64) {
	this.locker.Lock()
	defer this.locker.Unlock()

	item, itemOk := this.itemMap[key]
	if itemOk {
		item.Reset()
	} else {
		item = NewSysLockerIncrementItem(this.size)
		this.itemMap[key] = item
	}
	for i := minValue; i <= maxValue; i++ {
		item.Push(i)
	}
}

func (this *SysLockerIncrement) MaxValue(key string) int64 {
	this.locker.RLock()
	defer this.locker.RUnlock()

	item, itemOk := this.itemMap[key]
	if itemOk {
		return item.MaxValue()
	}
	return 0
}
