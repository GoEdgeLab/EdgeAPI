package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestMetricSumStatDAO_FindNodeSum(t *testing.T) {
	t.Log(models.NewMetricSumStatDAO().FindNodeSum(nil, 46, timeutil.Format("Ymd"), 1, 1))
}

func TestMetricSumStatDAO_Clean(t *testing.T) {
	dbs.NotifyReady()

	err := models.NewMetricSumStatDAO().Clean(nil)
	if err != nil {
		t.Fatal(err)
	}
}
