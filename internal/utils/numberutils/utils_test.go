// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package numberutils_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"testing"
)

func TestMax(t *testing.T) {
	t.Log(numberutils.Max[int](1, 2, 3))
	t.Log(numberutils.Max[int32](1, 2, 3))
	t.Log(numberutils.Max[float32](1.2, 2.3, 3.4))
}

func TestMin(t *testing.T) {
	t.Log(numberutils.Min[int](1, 2, 3))
	t.Log(numberutils.Min[int32](1, 2, 3))
	t.Log(numberutils.Min[float32](1.2, 2.3, 3.4))
}
