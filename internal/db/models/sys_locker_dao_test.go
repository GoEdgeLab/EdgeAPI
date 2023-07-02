package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"sync"
	"testing"
	"time"
)

func TestSysLockerDAO_Lock(t *testing.T) {
	var tx *dbs.Tx

	isOk, err := SharedSysLockerDAO.Lock(tx, "test", 600)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(isOk)

	if isOk {
		err = SharedSysLockerDAO.Unlock(tx, "test")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSysLocker_Increase(t *testing.T) {
	dbs.NotifyReady()

	var count = 1000

	var dao = NewSysLockerDAO()
	value, err := dao.Read(nil, "hello")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("before", value)

	var locker = sync.Mutex{}
	var allValueMap = map[int64]bool{}

	var before = time.Now()

	var wg = sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()

			var key = "hello"
			v, err := dao.Increase(nil, key, 0)
			if err != nil {
				t.Log("err:", err)
				return
			}

			locker.Lock()
			if allValueMap[v] {
				t.Log("duplicated:", v)
			} else {
				allValueMap[v] = true
			}
			locker.Unlock()

			//t.Log("v:", v)
			_ = v
		}(i)
	}

	wg.Wait()

	t.Log("cost:", time.Since(before).Seconds()*1000, "ms")

	value, err = dao.Read(nil, "hello")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("after", value, "values:", len(allValueMap))
}

func TestSysLocker_Increase_Performance(t *testing.T) {
	dbs.NotifyReady()

	var count = 1000

	var dao = NewSysLockerDAO()

	var before = time.Now()

	var wg = sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()

			var key = "hello" + types.String(i%10)
			v, err := dao.Increase(nil, key, 0)
			if err != nil {
				t.Log("err:", err)
				return
			}
			//t.Log("v:", v)
			_ = v
		}(i)
	}

	wg.Wait()

	t.Log("cost:", time.Since(before).Seconds()*1000, "ms")
}
