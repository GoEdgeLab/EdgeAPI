// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package utils_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"testing"
)

func TestLookupCNAME(t *testing.T) {
	t.Log(utils.LookupCNAME("www.yun4s.cn"))
}

func TestLookupNS(t *testing.T) {
	t.Log(utils.LookupNS("goedge.cn"))
}

func TestLookupTXT(t *testing.T) {
	t.Log(utils.LookupTXT("yanzheng.goedge.cn"))
}
