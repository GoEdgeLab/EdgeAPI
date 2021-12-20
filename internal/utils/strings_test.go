// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import (
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestSplitStrings(t *testing.T) {
	t.Log(SplitStrings("a, b, c", ","))
	t.Log(SplitStrings("a,      b, c, ", ","))
}

func TestContainsStringInsensitive(t *testing.T) {
	var a = assert.NewAssertion(t)
	a.IsTrue(ContainsStringInsensitive([]string{"a", "b", "C"}, "A"))
	a.IsTrue(ContainsStringInsensitive([]string{"a", "b", "C"}, "b"))
	a.IsTrue(ContainsStringInsensitive([]string{"a", "b", "C"}, "c"))
	a.IsFalse(ContainsStringInsensitive([]string{"a", "b", "C"}, "d"))

}
