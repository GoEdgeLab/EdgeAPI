// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils

import "testing"

func TestSplitStrings(t *testing.T) {
	t.Log(SplitStrings("a, b, c", ","))
	t.Log(SplitStrings("a,      b, c, ", ","))
}
