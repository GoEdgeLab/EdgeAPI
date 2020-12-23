package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestDNSDomainDAO_ExistDomainRecord(t *testing.T) {
	{
		b, err := NewDNSDomainDAO().ExistDomainRecord(1, "mycluster", "A", "", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)
	}
	{
		b, err := NewDNSDomainDAO().ExistDomainRecord(2, "mycluster", "A", "", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)
	}
	{
		b, err := NewDNSDomainDAO().ExistDomainRecord(2, "mycluster", "MX", "", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)
	}
	{
		b, err := NewDNSDomainDAO().ExistDomainRecord(2, "mycluster123", "A", "", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)
	}
}
