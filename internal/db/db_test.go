package db

import (
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestDB(t *testing.T) {
	Tea.Env = "prod"
	t.Log(dbs.Default())
}
