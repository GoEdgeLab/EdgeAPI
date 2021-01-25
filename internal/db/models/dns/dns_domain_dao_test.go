package dns

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestDNSDomainDAO_ExistDomainRecord(t *testing.T) {
	var tx *dbs.Tx

	{
		b, err := NewDNSDomainDAO().ExistDomainRecord(tx, 1, "mycluster", "A", "", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)
	}
	{
		b, err := NewDNSDomainDAO().ExistDomainRecord(tx, 2, "mycluster", "A", "", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)
	}
	{
		b, err := NewDNSDomainDAO().ExistDomainRecord(tx, 2, "mycluster", "MX", "", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)
	}
	{
		b, err := NewDNSDomainDAO().ExistDomainRecord(tx, 2, "mycluster123", "A", "", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)
	}
}
