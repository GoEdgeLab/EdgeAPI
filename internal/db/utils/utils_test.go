// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbutils_test

import (
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/iwind/TeaGo/assert"
	"testing"
)

func TestQuoteLike(t *testing.T) {
	for _, s := range []string{"abc", "abc%", "_abc%%%"} {
		t.Log(s + " => " + dbutils.QuoteLike(s))
	}
}

func TestIsLocalAddr(t *testing.T) {
	var a = assert.NewAssertion(t)
	a.IsTrue(dbutils.IsLocalAddr("127.0.0.1"))
	a.IsTrue(dbutils.IsLocalAddr("localhost"))
	a.IsTrue(dbutils.IsLocalAddr("::1"))
	a.IsTrue(dbutils.IsLocalAddr("127.0.0.1:3306"))
	a.IsFalse(dbutils.IsLocalAddr("192.168.2.200"))
	a.IsFalse(dbutils.IsLocalAddr("192.168.2.200:3306"))
}

func TestMySQLVersion(t *testing.T) {
	version, err := dbutils.MySQLVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("version:", version)
}

func TestMySQLVersionFrom8(t *testing.T) {
	t.Log(dbutils.MySQLVersionFrom8())
}

func TestFindMySQLPath(t *testing.T) {
	t.Log(dbutils.FindMySQLPath())
}

func TestStartLocalMySQL(t *testing.T) {
	dbutils.StartLocalMySQL()
}