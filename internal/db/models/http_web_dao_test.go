package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestHTTPWebDAO_UpdateWebShutdown(t *testing.T) {
	{
		err := SharedHTTPWebDAO.UpdateWebShutdown(1, []byte("{}"))
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		err := SharedHTTPWebDAO.UpdateWebShutdown(1, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Log("ok")
}
