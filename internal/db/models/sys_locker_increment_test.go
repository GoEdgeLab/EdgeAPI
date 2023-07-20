// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"testing"
)

func TestNewSysLockerIncrement(t *testing.T) {
	var increment = models.NewSysLockerIncrement(10)
	increment.Push("key", 1, 10)
	t.Log(increment.MaxValue("key"))
	for i := 0; i < 11; i++ {
		result, value := increment.Pop("key")
		t.Log(i, "=>", result, value)
	}

	for i := 0; i < 11; i++ {
		result, value := increment.Pop("key1")
		t.Log(i, "=>", result, value)
	}
}
