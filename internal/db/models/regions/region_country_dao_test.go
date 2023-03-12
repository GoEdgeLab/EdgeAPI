package regions

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestRegionCountryDAO_FindCountryIdWithName(t *testing.T) {
	dbs.NotifyReady()

	for _, name := range []string{
		"中国",
		"中华人民共和国",
		"美国",
		"美利坚合众国",
		"美利坚",
	} {
		countryId, err := SharedRegionCountryDAO.FindCountryIdWithName(nil, name)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(name, ":", countryId)
	}
}

func TestRegionCountryDAO_FindSimilarCountries(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	countries, err := SharedRegionCountryDAO.FindAllCountries(tx)
	if err != nil {
		t.Fatal(err)
	}

	for _, countryName := range []string{"中国", "布基纳法索", "哥伦比亚", "德意志共和国", "美利坚", "刚果金"} {
		t.Log("====" + countryName + "====")
		var countries = SharedRegionCountryDAO.FindSimilarCountries(countries, countryName, 5)
		if err != nil {
			t.Fatal(err)
		}
		for _, country := range countries {
			t.Log(country.Name, country.AllCodes())
		}
	}
}
