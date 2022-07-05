package utils_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"testing"
	"time"
)

func TestRangeDays(t *testing.T) {
	days, err := utils.RangeDays("20210101", "20210115")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(days)
}

func TestRangeMonth(t *testing.T) {
	days, err := utils.RangeMonths("20200101", "20210115")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(days)
}

func TestRangeHours(t *testing.T) {
	{
		hours, err := utils.RangeHours("2021010100", "2021010123")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(hours)
	}

	{
		hours, err := utils.RangeHours("2021010105", "2021010112")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(hours)
	}
}

func TestRangeMinutes(t *testing.T) {
	{
		var minutes = utils.RangeMinutes(time.Now(), 5, 5)
		t.Log(minutes)
	}

	{
		var minutes = utils.RangeMinutes(time.Now(), 5, 3)
		t.Log(minutes)
	}

	{
		var now = time.Now()
		var hour = now.Hour()
		var minute = now.Minute()
		now = now.Add(-time.Duration(hour) * time.Hour)
		now = now.Add(-time.Duration(minute-7) * time.Minute) // 后一天的 00:07 开始往前计算
		var minutes = utils.RangeMinutes(now, 5, 5)
		t.Log(minutes)
	}
}

func TestGroupMinuteRanges(t *testing.T) {
	{
		var minutes = utils.GroupMinuteRanges(utils.RangeMinutes(time.Now(), 5, 5))
		t.Log(minutes)
	}

	{
		var now = time.Now()
		var hour = now.Hour()
		var minute = now.Minute()
		now = now.Add(-time.Duration(hour) * time.Hour)
		now = now.Add(-time.Duration(minute-7) * time.Minute) // 后一天的 00:07 开始往前计算
		var minutes = utils.GroupMinuteRanges(utils.RangeMinutes(now, 5, 5))
		t.Log(minutes)
	}
}
