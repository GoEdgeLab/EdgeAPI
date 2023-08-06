// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package utils_test

import (
	"errors"
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

type jsonUserType struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (this *jsonUserType) Init() error {
	if len(this.Name) < 10 {
		return errors.New("'name' too short")
	}
	return nil
}

func TestJSONDecodeConfig(t *testing.T) {
	var data = []byte(`{ "name":"Lily", "age":20, "description": "Nice" }`)

	var u = &jsonUserType{}
	newJSON, err := utils.JSONDecodeConfig(data, u)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v, %s", u, string(newJSON))
}

func TestJSONDecodeConfig_Validate(t *testing.T) {
	var data = []byte(`{ "name":"Lily", "age":20, "description": "Nice" }`)

	var u = &jsonUserType{}

	newJSON, err := utils.JSONDecodeConfig(data, u)
	if err != nil {
		t.Log("ignore error:", err) // error expected
	}
	t.Logf("%+v, %s", u, string(newJSON))
}
