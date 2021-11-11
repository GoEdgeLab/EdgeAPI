// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import (
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestNewCacheMap(t *testing.T) {
	var a = assert.NewAssertion(t)

	m := NewCacheMap()
	{
		m.Put("Hello", "World")
		v, ok := m.Get("Hello")
		a.IsTrue(ok)
		a.IsTrue(v == "World")
	}

	{
		v, ok := m.Get("Hello1")
		a.IsFalse(ok)
		a.IsTrue(v == nil)
	}
}
