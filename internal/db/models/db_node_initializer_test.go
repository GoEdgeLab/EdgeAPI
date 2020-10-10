package models

import (
	"runtime"
	"testing"
	"time"
)

func TestDBNodeInitializer_loop(t *testing.T) {
	initializer := NewDBNodeInitializer()
	err := initializer.loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(accessLogDBMapping), len(accessLogDAOMapping))
}

func TestFindAccessLogTable(t *testing.T) {
	before := time.Now()
	db := SharedHTTPAccessLogDAO.Instance
	tableName, err := findAccessLogTable(db, "20201010", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tableName)
	t.Log(time.Since(before).Seconds()*1000, "ms")

	before = time.Now()
	tableName, err = findAccessLogTable(db, "20201010", false)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(tableName)
	t.Log(time.Since(before).Seconds()*1000, "ms")
}

func BenchmarkFindAccessLogTable(b *testing.B) {
	db := SharedHTTPAccessLogDAO.Instance

	runtime.GOMAXPROCS(1)
	for i := 0; i < b.N; i++ {
		_, _ = findAccessLogTable(db, "20201010", false)
	}
}
