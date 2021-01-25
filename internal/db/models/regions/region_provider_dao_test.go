package regions

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestRegionProviderDAO_FindProviderIdWithProviderNameCacheable(t *testing.T) {
	dbs.NotifyReady()

	for i := 0; i < 5; i++ {
		now := time.Now()
		providerId, err := SharedRegionProviderDAO.FindProviderIdWithNameCacheable(nil, "电信")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("providerId", providerId, time.Since(now).Seconds()*1000, "ms")
	}

	t.Log("=====")

	for i := 0; i < 5; i++ {
		now := time.Now()
		providerId, err := SharedRegionProviderDAO.FindProviderIdWithNameCacheable(nil, "胡乱填的")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("providerId", providerId, time.Since(now).Seconds()*1000, "ms")
	}
}
