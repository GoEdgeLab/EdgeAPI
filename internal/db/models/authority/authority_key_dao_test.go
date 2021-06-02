package authority

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestAuthorityKeyDAO_UpdateValue(t *testing.T) {
	err := NewAuthorityKeyDAO().UpdateKey(nil, "12345678", "", "", "", []string{}, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestAuthorityKeyDAO_ReadValue(t *testing.T) {
	value, err := NewAuthorityKeyDAO().ReadKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(value)
}
