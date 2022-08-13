package regions

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestRegionProvinceDAO_FindProvinceIdWithNameCacheable(t *testing.T) {
	dbs.NotifyReady()

	for i := 0; i < 5; i++ {
		now := time.Now()
		provinceId, err := SharedRegionProvinceDAO.FindProvinceIdWithNameCacheable(nil, 1, "安徽省")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(provinceId, time.Since(now).Seconds()*1000, "ms")
	}

	t.Log("====")
	for i := 0; i < 5; i++ {
		now := time.Now()
		provinceId, err := SharedRegionProvinceDAO.FindProvinceIdWithNameCacheable(nil, 2, "安徽省")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(provinceId, time.Since(now).Seconds()*1000, "ms")
	}
}

func TestRegionProvinceDAO_FindProvinceIdWithName(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	for _, name := range []string{
		"安徽",
		"安徽省",
		"广西",
		"广西省",
		"广西壮族自治区",
		"皖",
	} {
		provinceId, err := SharedRegionProvinceDAO.FindProvinceIdWithName(tx, 1, name)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(name, "=>", provinceId)
	}
}

func TestRegionProvinceDAO_FindSimilarProvinces(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	var countryId int64 = 1
	provinces, err := SharedRegionProvinceDAO.FindAllEnabledProvincesWithCountryId(tx, countryId)
	if err != nil {
		t.Fatal(err)
	}

	for _, provinceName := range []string{
		"北京",
		"北京市",
		"安徽",
		"安徽省",
		"大北京",
	} {
		t.Log("====" + provinceName + "====")
		var provinces = SharedRegionProvinceDAO.FindSimilarProvinces(provinces, provinceName, 5)
		if err != nil {
			t.Fatal(err)
		}
		for _, province := range provinces {
			t.Log(province.Name, province.AllCodes())
		}
	}
}
