package utils

import "testing"

func TestRangeDays(t *testing.T) {
	days, err := RangeDays("20210101", "20210115")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(days)
}


func TestRangeHours(t *testing.T) {
	{
		hours, err := RangeHours("2021010100", "2021010123")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(hours)
	}

	{
		hours, err := RangeHours("2021010105", "2021010112")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(hours)
	}
}
