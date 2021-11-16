package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"sync"
	"testing"
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
	count := 100
	wg := sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			v, err := NewSysLockerDAO().Increase(nil, "hello", 0)
			if err != nil {
				t.Log("err:", err)
				return
			}
			t.Log("v:", v)
		}()
	}

	wg.Wait()
	t.Log("ok")
}
