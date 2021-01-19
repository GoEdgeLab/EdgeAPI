package db

import (
	"database/sql"
	"database/sql/driver"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestDB_Env(t *testing.T) {
	Tea.Env = "prod"
	t.Log(dbs.Default())
}

func TestDB_Instance(t *testing.T) {
	for i := 0; i < 10; i++ {
		db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s")
		if err != nil {
			t.Fatal(i, "open:", err)
		}
		db.SetConnMaxIdleTime(time.Minute * 3)
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxIdleConns(0)
		db.SetMaxOpenConns(1)
		go func(db *sql.DB, i int) {
			for j := 0; j < 100; j++ {
				err := db.Ping()
				if err != nil {
					if err == driver.ErrBadConn {
						return
					}
					t.Fatal(i, "exec:", err)
				}
				time.Sleep(1 * time.Second)
			}
			t.Log(i, "ok", db)
		}(db, i)
	}
	time.Sleep(100 * time.Second)
}
