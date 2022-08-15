// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestSplitStrings(t *testing.T) {
	t.Log(utils.SplitStrings("a, b, c", ","))
	t.Log(utils.SplitStrings("a,      b, c, ", ","))
}

func TestContainsStringInsensitive(t *testing.T) {
	var a = assert.NewAssertion(t)
	a.IsTrue(utils.ContainsStringInsensitive([]string{"a", "b", "C"}, "A"))
	a.IsTrue(utils.ContainsStringInsensitive([]string{"a", "b", "C"}, "b"))
	a.IsTrue(utils.ContainsStringInsensitive([]string{"a", "b", "C"}, "c"))
	a.IsFalse(utils.ContainsStringInsensitive([]string{"a", "b", "C"}, "d"))
}

func TestSimilar(t *testing.T) {
	t.Log(utils.Similar("", ""))
	t.Log(utils.Similar("", "a"))
	t.Log(utils.Similar("abc", "bcd"))
	t.Log(utils.Similar("efgj", "hijk"))
	t.Log(utils.Similar("efgj", "klmn"))
}

func TestLimitString(t *testing.T) {
	var a = assert.NewAssertion(t)
	a.IsTrue(utils.LimitString("", 4) == "")
	a.IsTrue(utils.LimitString("abcd", 0) == "")
	a.IsTrue(utils.LimitString("abcd", 5) == "abcd")
	a.IsTrue(utils.LimitString("abcd", 4) == "abcd")
	a.IsTrue(utils.LimitString("abcd", 3) == "abc")
	a.IsTrue(utils.LimitString("abcd", 1) == "a")
	a.IsTrue(utils.LimitString("中文测试", 1) == "")
	a.IsTrue(utils.LimitString("中文测试", 3) == "中")
}
