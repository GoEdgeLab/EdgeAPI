package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestOriginServerDAO_ComposeOriginConfig(t *testing.T) {
	config, err := SharedOriginDAO.ComposeOriginConfig(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
}
