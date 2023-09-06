package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestUserPlanBandwidthStatDAO_FindMonthlyPercentile(t *testing.T) {
	var dao = models.NewUserPlanBandwidthStatDAO()
	var tx *dbs.Tx

	{
		resultBytes, err := dao.FindMonthlyPercentile(tx, 20, timeutil.Format("Ym"), 100, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("result bytes0:", resultBytes)
	}

	{
		resultBytes, err := dao.FindMonthlyPercentile(tx, 20, timeutil.Format("Ym"), 95, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("result bytes1:", resultBytes)
	}

	{
		resultBytes, err := dao.FindMonthlyPercentile(tx, 20, timeutil.Format("Ym"), 95, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("result bytes2:", resultBytes)
	}
}
