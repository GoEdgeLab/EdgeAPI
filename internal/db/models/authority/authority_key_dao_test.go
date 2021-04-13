package authority

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestAuthorityKeyDAO_UpdateValue(t *testing.T) {
	err := NewAuthorityKeyDAO().UpdateValue(nil, "12345678")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ok")
}

func TestAuthorityKeyDAO_ReadValue(t *testing.T) {
	value, err := NewAuthorityKeyDAO().ReadValue(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(value)
}
