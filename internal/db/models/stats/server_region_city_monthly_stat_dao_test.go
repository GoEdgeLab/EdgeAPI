package stats

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestServerRegionCityMonthlyStatDAO_Clean(t *testing.T) {
	var dao = NewServerRegionCityMonthlyStatDAO()
	err := dao.Clean(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
