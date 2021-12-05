package stats

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestServerClientSystemMonthlyStatDAO_Clean(t *testing.T) {
	var dao = NewServerClientSystemMonthlyStatDAO()
	err := dao.Clean(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
