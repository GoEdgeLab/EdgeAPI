package regions

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestRegionCountryDAO_FindCountryIdWithCountryNameCacheable(t *testing.T) {
	dbs.NotifyReady()

	for i := 0; i < 5; i++ {
		now := time.Now()
		countryId, err := SharedRegionCountryDAO.FindCountryIdWithNameCacheable(nil, "中国")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("countryId", countryId, time.Since(now).Seconds()*1000, "ms")
	}
}
