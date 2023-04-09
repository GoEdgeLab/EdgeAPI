// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestJSONClone(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}

	var u = &user{
		Name: "Jack",
		Age:  20,
	}

	newU, err := utils.JSONClone[*user](u)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", newU)
}

func TestJSONClone_Slice(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}

	var u = []*user{
		{
			Name: "Jack",
			Age:  20,
		},
		{
			Name: "Lily",
			Age:  18,
		},
	}

	newU, err := utils.JSONClone[[]*user](u)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(newU, t)
}
