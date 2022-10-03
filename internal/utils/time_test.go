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

func TestRangeTimes(t *testing.T) {
	for _, r := range [][2]string{
		{"0000", "2359"},
		{"0000", "0230"},
		{"0300", "0230"},
		{"1021", "1131"},
	} {
		result, err := utils.RangeTimes(r[0], r[1], 5)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(r, "=>", result, len(result))
	}
}

func TestRange24HourTimes(t *testing.T) {
	t.Log(utils.Range24HourTimes(5))
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
