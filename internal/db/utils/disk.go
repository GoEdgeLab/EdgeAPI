// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dbutils

import (
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"golang.org/x/sys/unix"
	"time"
)

const minFreeSpaceGB = 3

var HasFreeSpace = true
var IsLocalDatabase = false
var LocalDatabaseDataDir = ""

func init() {
	var ticker = time.NewTicker(5 * time.Minute)

	dbs.OnReadyDone(func() {
		goman.New(func() {
			for range ticker.C {
				HasFreeSpace = CheckHasFreeSpace()
			}
		})
	})
}

// CheckHasFreeSpace 检查当前数据库是否有剩余空间
func CheckHasFreeSpace() bool {
	db, _ := dbs.Default()
	if db == nil {
		return false
	}

	config, _ := db.Config()
	if config == nil {
		return false
	}

	dsnConfig, _ := mysql.ParseDSN(config.Dsn)
	if dsnConfig == nil {
		return false
	}

	if IsLocalAddr(dsnConfig.Addr) {
		IsLocalDatabase = true

		// only for local database
		one, err := db.FindOne("SHOW VARIABLES WHERE variable_name='datadir'")
		if err != nil || len(one) == 0 {
			return true
		}

		var dir = one.GetString("Value")
		if len(dir) == 0 {
			return true
		}
		LocalDatabaseDataDir = dir

		var stat unix.Statfs_t
		err = unix.Statfs(dir, &stat)
		if err != nil {
			return true
		}

		var availableSpace = (stat.Bavail * uint64(stat.Bsize)) / (1 << 30) // GB
		return availableSpace > minFreeSpaceGB
	}
	return true
}
