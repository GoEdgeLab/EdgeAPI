// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbutils_test

import (
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"testing"
)

func TestQuoteLike(t *testing.T) {
	for _, s := range []string{"abc", "abc%", "_abc%%%"} {
		t.Log(s + " => " + dbutils.QuoteLike(s))
	}
}
