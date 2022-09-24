// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"testing"
)

func TestEqualConfig(t *testing.T) {
	type testType struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	{
		var c1 = &testType{
			Name: "Lily",
			Age:  12,
		}
		var c2 = &testType{
			Name: "Lucy",
			Age:  12,
		}
		t.Log(utils.EqualConfig(c1, c2))
	}

	{
		var c1 = &testType{
			Name: "Lily",
			Age:  12,
		}
		var c2 = &testType{
			Age:  12,
			Name: "Lily",
		}
		t.Log(utils.EqualConfig(c1, c2))
	}
}
