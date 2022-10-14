// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package numberutils_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"math"
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

func TestMaxFloat32(t *testing.T) {
	t.Logf("%f", math.MaxFloat32/(1<<100))
}

func TestFloorFloat32(t *testing.T) {
	t.Logf("%f", numberutils.FloorFloat32(123.456, -1))
	t.Logf("%f", numberutils.FloorFloat32(123.456, 0))
	t.Logf("%f, %f", numberutils.FloorFloat32(123.456, 1), 123.456*10)
	t.Logf("%f, %f", numberutils.FloorFloat32(123.456, 2), 123.456*10*10)
	t.Logf("%f, %f", numberutils.FloorFloat32(123.456, 3), 123.456*10*10*10)
	t.Logf("%f, %f", numberutils.FloorFloat32(123.456, 4), 123.456*10*10*10*10)
}
