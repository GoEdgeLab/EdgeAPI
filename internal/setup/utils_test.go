// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package setup_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestComposeSQLVersion(t *testing.T) {
	t.Log(setup.ComposeSQLVersion())
}

func TestCompareVersion(t *testing.T) {
	var a = assert.NewAssertion(t)
	a.IsTrue(setup.CompareVersion("1.3.4", "1.3.4") == 0)
	a.IsTrue(setup.CompareVersion("1.3.4", "1.3.3") > 0)
	a.IsTrue(setup.CompareVersion("1.3.4", "1.3.5") < 0)
	a.IsTrue(setup.CompareVersion("1.3.4.3", "1.3.4.12") > 0) // because 12 > 10
	a.IsTrue(setup.CompareVersion("1.3.4.3", "1.3.4.2") > 0)
	a.IsTrue(setup.CompareVersion("1.3.4.3", "1.3.4.4") < 0)
}
