// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package models_test

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/assert"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestIsMySQLError(t *testing.T) {
	var a = assert.NewAssertion(t)

	{
		var err error
		a.IsFalse(models.IsMySQLError(err))
	}

	{
		var err = errors.New("hello")
		a.IsFalse(models.IsMySQLError(err))
	}

	{
		db, err := dbs.Default()
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = db.Close()
		}()
		_, err = db.Exec("SELECT abc")
		a.IsTrue(models.IsMySQLError(err))
		a.IsTrue(models.CheckSQLErrCode(err, 1054))
		a.IsFalse(models.CheckSQLErrCode(err, 1000))
	}
}
